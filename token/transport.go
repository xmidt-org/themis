package token

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/xmidt-org/themis/xhttp/xhttpserver"

	"github.com/gorilla/mux"
	"go.uber.org/multierr"
)

var (
	ErrVariableNotAllowed = errors.New("Either header/parameter or variable can specified, but not all three")
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
		if sc, ok := err.(kithttp.StatusCoder); ok {
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
	}

	return nil
}

// NewRequestBuilders creates a RequestBuilders sequence given an Options configuration.  Only claims
// and metadata that are HTTP-based are included in the results.  Claims and metadata that are statically
// assigned values are handled by ClaimBuilder objects and are part of the Factory configuration.
func NewRequestBuilders(o Options) (RequestBuilders, error) {
	var rb RequestBuilders
	for name, value := range o.Claims {
		if len(value.Header) > 0 || len(value.Parameter) > 0 {
			if len(value.Variable) > 0 {
				return nil, ErrVariableNotAllowed
			}

			rb = append(rb,
				headerParameterRequestBuilder{
					key:       name,
					header:    http.CanonicalHeaderKey(value.Header),
					parameter: value.Parameter,
					setter:    claimsSetter,
				},
			)
		} else if len(value.Variable) > 0 {
			rb = append(rb,
				variableRequestBuilder{
					key:      name,
					variable: value.Variable,
					setter:   claimsSetter,
				},
			)
		}
	}

	for name, value := range o.Metadata {
		if len(value.Header) > 0 || len(value.Parameter) > 0 {
			if len(value.Variable) > 0 {
				return nil, ErrVariableNotAllowed
			}

			rb = append(rb,
				headerParameterRequestBuilder{
					key:       name,
					header:    http.CanonicalHeaderKey(value.Header),
					parameter: value.Parameter,
					setter:    metadataSetter,
				},
			)
		} else if len(value.Variable) > 0 {
			rb = append(rb,
				variableRequestBuilder{
					key:      name,
					variable: value.Variable,
					setter:   metadataSetter,
				},
			)
		}
	}

	if o.PartnerID != nil && (len(o.PartnerID.Claim) > 0 || len(o.PartnerID.Metadata) > 0) {
		rb = append(rb,
			partnerIDRequestBuilder{
				PartnerID: *o.PartnerID,
			},
		)
	}

	return rb, nil
}

// BuildRequest applies a sequence of RequestBuilder instances to produce a token factory Request
func BuildRequest(original *http.Request, rb RequestBuilders) (*Request, error) {
	tr := NewRequest()
	if err := rb.Build(original, tr); err != nil {
		return nil, err
	}

	return tr, nil
}

func DecodeServerRequest(rb RequestBuilders) func(context.Context, *http.Request) (interface{}, error) {
	return func(ctx context.Context, hr *http.Request) (interface{}, error) {
		if err := hr.ParseForm(); err != nil {
			return nil, err
		}

		tr, err := BuildRequest(hr, rb)
		if err != nil {
			return nil, err
		}

		return tr, nil
	}
}

func EncodeIssueResponse(_ context.Context, response http.ResponseWriter, value interface{}) error {
	response.Header().Set("Content-Type", "application/jose")
	_, err := response.Write([]byte(value.(string)))
	return err
}

type DecodeClaimsError struct {
	URL        string
	StatusCode int
	Err        error
}

func (dce *DecodeClaimsError) Unwrap() error {
	return dce.Err
}

func (dce *DecodeClaimsError) nestedErrorText() string {
	if dce.Err != nil {
		return dce.Err.Error()
	}

	return ""
}

func (dce *DecodeClaimsError) Error() string {
	return fmt.Sprintf(
		"Failed to decode remote claims from [%s]: statusCode=%d, err=%s",
		dce.URL,
		dce.StatusCode,
		dce.nestedErrorText(),
	)
}

func (dce *DecodeClaimsError) MarshalJSON() ([]byte, error) {
	var output bytes.Buffer
	fmt.Fprintf(
		&output,
		`{"url": "%s", "statusCode": %d, "err": "%s"}`,
		dce.URL,
		dce.StatusCode,
		dce.nestedErrorText(),
	)

	return output.Bytes(), nil
}

func DecodeRemoteClaimsResponse(_ context.Context, response *http.Response) (interface{}, error) {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode < 200 || response.StatusCode > 299 {
		err := &DecodeClaimsError{
			StatusCode: response.StatusCode,
		}

		if response.Request != nil {
			err.URL = response.Request.URL.String()
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
