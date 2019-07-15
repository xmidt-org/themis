package token

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	ErrNoHttp = errors.New("At least one of header, parameter, or variable must be specified")
)

type RequestBuilder func(*http.Request, *Request) error

type missingValueError struct {
	name  string
	value *Value
}

func (mve *missingValueError) Error() string {
	var output bytes.Buffer
	output.WriteString("Missing value '")
	output.WriteString(mve.name)
	output.WriteString("' from")

	first := true
	if len(mve.value.Header) > 0 {
		output.WriteString(" header '")
		output.WriteString(mve.value.Header)
		output.WriteString("'")
		first = false
	}

	if len(mve.value.Parameter) > 0 {
		if !first {
			output.WriteString(" or")
		}

		output.WriteString(" parameter '")
		output.WriteString(mve.value.Parameter)
		output.WriteString("'")
		first = false
	}

	if len(mve.value.Variable) > 0 {
		if !first {
			output.WriteString(" or")
		}

		output.WriteString(" variable '")
		output.WriteString(mve.value.Variable)
		output.WriteString("'")
		first = false
	}

	return output.String()
}

func (mve *missingValueError) StatusCode() int {
	return http.StatusBadRequest
}

func extractValue(original *http.Request, value *Value) (interface{}, bool) {
	if len(value.Header) > 0 {
		hv := original.Header.Get(value.Header)
		if len(hv) > 0 {
			return hv, true
		}
	}

	if len(value.Parameter) > 0 {
		pv := original.Form[value.Parameter]
		if len(pv) > 0 {
			return pv[0], true
		}
	}

	if len(value.Variable) > 0 {
		vv := mux.Vars(original)[value.Variable]
		if len(vv) > 0 {
			return vv, true
		}
	}

	return nil, false
}

func NewClaimBuilder(name string, value Value) (RequestBuilder, error) {
	if !value.IsHttp() {
		return nil, ErrNoHttp
	}

	mve := missingValueError{name: name, value: &value}
	return func(original *http.Request, tr *Request) error {
		if v, ok := extractValue(original, &value); ok {
			tr.Claims[name] = v
			return nil
		}

		return &mve
	}, nil
}

func NewMetadataBuilder(name string, value Value) (RequestBuilder, error) {
	if !value.IsHttp() {
		return nil, ErrNoHttp
	}

	mve := missingValueError{name: name, value: &value}
	return func(original *http.Request, tr *Request) error {
		if v, ok := extractValue(original, &value); ok {
			tr.Metadata[name] = v
			return nil
		}

		return &mve
	}, nil
}

func NewTokenRequestBuilders(o Options) []RequestBuilder {
	var builders []RequestBuilder

	for name, value := range o.Claims {
		if b, err := NewClaimBuilder(name, value); err == nil {
			builders = append(builders, b)
		}
	}

	for name, value := range o.Metadata {
		if b, err := NewMetadataBuilder(name, value); err == nil {
			builders = append(builders, b)
		}
	}

	return builders
}

func BuildRequest(hr *http.Request, b []RequestBuilder) (*Request, error) {
	tr := &Request{
		Claims:   make(map[string]interface{}),
		Metadata: make(map[string]interface{}),
	}

	for _, f := range b {
		if err := f(hr, tr); err != nil {
			return nil, err
		}
	}

	return tr, nil
}

func DecodeServerRequest(b ...RequestBuilder) func(context.Context, *http.Request) (interface{}, error) {
	return func(_ context.Context, hr *http.Request) (interface{}, error) {
		tr, err := BuildRequest(hr, b)
		if err != nil {
			return nil, err
		}

		return tr, nil
	}
}

func EncodeServerResponse(_ context.Context, response http.ResponseWriter, value interface{}) error {
	response.Header().Set("Content-Type", "application/jose")
	_, err := response.Write([]byte(value.(string)))
	return err
}

func DecodeRemoteClaimsResponse(_ context.Context, response *http.Response) (interface{}, error) {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(body, &claims); err != nil {
		return nil, err
	}

	return claims, nil
}
