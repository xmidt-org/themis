// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package token

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"slices"
	"strings"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/xmidt-org/sallust"
	"github.com/xmidt-org/themis/xhttp/xhttpserver"

	"github.com/gorilla/mux"
	"go.uber.org/multierr"
)

var (
	ErrVariableNotAllowed = errors.New("either header/parameter or variable can specified, but not all three")
)

// InvalidPartnerIDError is the error object returned when a blank, wildcard, or otherwise
// invalid partner id is submitted
type InvalidPartnerIDError struct{}

// Error returns the error string associated with an invalid partner id
func (ipe InvalidPartnerIDError) Error() string {
	return "invalid partner id"
}

func (ipe InvalidPartnerIDError) StatusCode() int {
	return http.StatusBadRequest
}

// BuildError is the error type usually returned by RequestBuilder.Build to indicate what
// happened during each request builder.
type BuildError struct {
	Err error
}

// Error returns the nested error's Error text
func (be BuildError) Error() string {
	return be.Err.Error()
}

func (be BuildError) Unwrap() error {
	return be.Err
}

// StatusCode returns the largest numeric HTTP status code of any embedded errors,
// or http.StatusBadRequest is none of the embedded errors reported status codes.
func (be BuildError) StatusCode() int {
	statusCode := 0
	for _, err := range multierr.Errors(be.Err) {

		var sc kithttp.StatusCoder
		if ok := errors.As(err, &sc); ok {
			if statusCode < sc.StatusCode() {
				statusCode = sc.StatusCode()
			}
		}
	}

	if statusCode == 0 {
		return http.StatusBadRequest
	}

	return statusCode
}

// RequestBuilder is a strategy for building a token factory Request from an HTTP request.
//
// Note: before invoking a RequestBuilder, calling code should parse the HTTP request form.
type RequestBuilder interface {
	Build(*http.Request, *Request) error
}

type RequestBuilderFunc func(*http.Request, *Request) error

func (rbf RequestBuilderFunc) Build(original *http.Request, tr *Request) error {
	return rbf(original, tr)
}

// RequestBuilders represents a set of RequestBuilder strategies that can be invoked in sequence.
type RequestBuilders []RequestBuilder

// Build invokes each request builder in sequence.  Prior to invoking any of the chain of builders,
func (rbs RequestBuilders) Build(original *http.Request, tr *Request) error {
	var err error
	for _, rb := range rbs {
		multierr.AppendInto(&err, rb.Build(original, tr))
	}

	if err != nil {
		return BuildError{Err: err}
	}

	return nil
}

func claimsSetter(key string, value interface{}, tr *Request) {
	tr.Claims[key] = value
}

func metadataSetter(key string, value interface{}, tr *Request) {
	tr.Metadata[key] = value
}

func pathWildCardsSetter(key string, value interface{}, tr *Request) {
	tr.PathWildCards[key] = value
}

func queryParametersSetter(key string, value interface{}, tr *Request) {
	tr.QueryParameters[key] = value
}

type headerParameterRequestBuilder struct {
	key       string
	header    string
	parameter string
	setter    func(string, interface{}, *Request)
}

func (hprb headerParameterRequestBuilder) Build(original *http.Request, tr *Request) error {
	if len(hprb.header) > 0 {
		value := original.Header[hprb.header]
		if len(value) > 0 {
			hprb.setter(hprb.key, value[0], tr)
			return nil
		}
	}

	if len(hprb.parameter) > 0 {
		value := original.Form[hprb.parameter]
		if len(value) > 0 {
			hprb.setter(hprb.key, value[0], tr)
			return nil
		}
	}

	return nil
}

type variableRequestBuilder struct {
	key      string
	variable string
	setter   func(string, interface{}, *Request)
}

func (vrb variableRequestBuilder) Build(original *http.Request, tr *Request) error {
	value := mux.Vars(original)[vrb.variable]
	if len(value) > 0 {
		vrb.setter(vrb.key, value, tr)
		return nil
	}

	return xhttpserver.MissingVariableError{Variable: vrb.variable}
}

type staticRequestBuilder struct {
	key    string
	value  any
	setter func(string, interface{}, *Request)
}

func (srb staticRequestBuilder) Build(original *http.Request, tr *Request) error {
	srb.setter(srb.key, srb.value, tr)

	return nil
}

type partnerIDRequestBuilder struct {
	PartnerID
}

func (prb partnerIDRequestBuilder) getPartnerID(original *http.Request) (string, error) {
	var value string
	if len(prb.Header) > 0 {
		value = original.Header.Get(prb.Header)
	}

	if len(value) == 0 && len(prb.Parameter) > 0 {
		values := original.Form[prb.Parameter]
		if len(values) > 0 {
			value = values[0]
		}
	}

	if len(value) > 0 {
		// some post-processing on the partner id value:
		// don't allow multiple values separated by ","
		// don't allow the "*" partner id
		for _, v := range strings.Split(value, ",") {
			v = strings.TrimSpace(v)
			if len(v) > 0 && v != "*" {
				return v, nil // the cleaned partner id
			}
		}

		// a partner id must have at least (1) segment that is not blank and is not the wildcard '*'
		return "", InvalidPartnerIDError{}
	}

	// return the default as is, without any of the special processing above
	return prb.Default, nil
}

func (prb partnerIDRequestBuilder) Build(original *http.Request, tr *Request) error {
	partnerID, err := prb.getPartnerID(original)
	if err != nil {
		return err
	}

	if len(partnerID) > 0 {
		if len(prb.Claim) > 0 {
			tr.Claims[prb.Claim] = partnerID
		}

		if len(prb.Metadata) > 0 {
			tr.Metadata[prb.Metadata] = partnerID
		}

		if len(prb.PathWildCard) > 0 {
			tr.PathWildCards[prb.PathWildCard] = partnerID
		}

		if len(prb.QueryParameter) > 0 {
			tr.QueryParameters[prb.QueryParameter] = partnerID
		}
	}

	return nil
}

// setConnectionState sets the tls.ConnectionState for the given request.
func setConnectionState(original *http.Request, tr *Request) error {
	if original.TLS != nil {
		tr.TLS = original.TLS
	}

	return nil
}

// NewRequestBuilders creates a RequestBuilders sequence given an Options configuration.  Only claims
// and metadata that are HTTP-based are included in the results.  Claims and metadata that are statically
// assigned values are handled by ClaimBuilder objects and are part of the Factory configuration.
func NewRequestBuilders(o Options) (rbs RequestBuilders, errs error) {
	rb, err := newRequestBuilders(o.Claims, claimsSetter)
	rb1, err1 := newRequestBuilders(o.Metadata, metadataSetter)
	rb2, err2 := newRequestStaticBuilders(o.PathWildCards, pathWildCardsSetter)
	rb3, err3 := newRequestStaticBuilders(o.QueryParameters, queryParametersSetter)
	rb4, err4 := newRequestBuilders(o.PathWildCards, pathWildCardsSetter)
	rb5, err5 := newRequestBuilders(o.QueryParameters, queryParametersSetter)

	if errs = errors.Join(err, err1, err2, err3, err4, err5); errs != nil {
		return nil, errs
	}

	rbs = slices.Concat(rb, rb1, rb2, rb3, rb4, rb5)
	if o.PartnerID != nil {
		rbs = append(rbs, partnerIDRequestBuilder{PartnerID: *o.PartnerID})
	}

	return append(rbs, RequestBuilderFunc(setConnectionState)), nil
}

func newRequestBuilders(values []Value, setter func(string, any, *Request)) (rbs RequestBuilders, errs error) {
	for _, v := range values {
		errs = multierr.Append(errs, v.Validate())
		if !v.IsFromHTTP() {
			continue
		}

		if len(v.Header) > 0 || len(v.Parameter) > 0 {
			rbs = append(rbs, headerParameterRequestBuilder{
				key:       v.Key,
				header:    http.CanonicalHeaderKey(v.Header),
				parameter: v.Parameter,
				setter:    setter,
			})
		} else {
			rbs = append(rbs, variableRequestBuilder{
				key:      v.Key,
				variable: v.Variable,
				setter:   setter,
			})
		}
	}

	return
}

func newRequestStaticBuilders(values []Value, setter func(string, any, *Request)) (rbs RequestBuilders, errs error) {
	for _, v := range values {
		errs = multierr.Append(errs, v.Validate())
		if !v.IsStatic() {
			continue
		}

		rbs = append(rbs, staticRequestBuilder{
			key:    v.Key,
			value:  v.Value,
			setter: setter,
		})
	}

	return
}

// BuildRequest applies a sequence of RequestBuilder instances to produce a token factory Request
func BuildRequest(original *http.Request, rb RequestBuilders) (*Request, error) {
	tr := NewRequest()
	if err := rb.Build(original, tr); err != nil {
		return nil, err
	}

	tr.Logger = sallust.Get(original.Context())
	return tr, nil
}

func DecodeServerRequest(rb RequestBuilders) func(context.Context, *http.Request) (interface{}, error) {
	return func(ctx context.Context, hr *http.Request) (interface{}, error) {
		if err := hr.ParseForm(); err != nil {
			return nil, httpError{
				err:  err,
				code: http.StatusBadRequest,
			}
		}

		tr, err := BuildRequest(hr, rb)
		if err != nil {
			return nil, err
		}

		tr.Logger = sallust.Get(ctx)
		return tr, nil
	}
}

func EncodeIssueResponse(_ context.Context, response http.ResponseWriter, value interface{}) error {
	response.Header().Set("Content-Type", "application/jose")
	_, err := response.Write([]byte(value.(string)))
	return err
}

type DecodeClaimsError struct {
	StatusCode int
	Err        error
}

func (dce DecodeClaimsError) Unwrap() error {
	return dce.Err
}

func (dce DecodeClaimsError) nestedErrorText() string {
	if dce.Err != nil {
		return dce.Err.Error()
	}

	return ""
}

func (dce DecodeClaimsError) Error() string {
	return fmt.Sprintf(
		"Failed to decode remote claims from: statusCode=%d, err=%s",
		dce.StatusCode,
		dce.nestedErrorText(),
	)
}

func (dce DecodeClaimsError) MarshalJSON() ([]byte, error) {
	var output bytes.Buffer
	fmt.Fprintf(
		&output,
		`{"statusCode": %d, "err": "%s"}`,
		dce.StatusCode,
		dce.nestedErrorText(),
	)

	return output.Bytes(), nil
}

func DecodeRemoteClaimsResponse(_ context.Context, response *http.Response) (interface{}, error) {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode < 200 || response.StatusCode > 299 {
		err := DecodeClaimsError{
			StatusCode: response.StatusCode,
		}

		if len(body) != 0 {
			err.Err = errors.New(string(body))
		}

		return nil, err
	}

	// allow empty bodies
	var claims map[string]interface{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &claims); err != nil {
			return nil, err
		}
	}

	return claims, nil
}

func EncodeRemoteClaimsRequest(c context.Context, r *http.Request, request interface{}) error {
	if headerer, ok := request.(kithttp.Headerer); ok {
		for k := range headerer.Headers() {
			r.Header.Set(k, headerer.Headers().Get(k))
		}
	}

	tr := request.(*Request)
	for k, v := range tr.PathWildCards {
		s, ok := v.(string)
		if !ok {
			return fmt.Errorf("remote claims expected a string path wild card value: %s", reflect.TypeOf(v))
		}

		r.URL.Path = strings.ReplaceAll(r.URL.Path, fmt.Sprintf("{%s}", k), s)
	}

	q := r.URL.Query()
	for k, v := range tr.QueryParameters {
		q.Add(k, v.(string))
	}

	r.URL.RawQuery = q.Encode()
	b, err := json.Marshal(tr.Metadata)
	if err != nil {
		return err
	}

	r.Body = io.NopCloser(bytes.NewReader(b))

	return nil
}
