package token

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
)

type RequestBuilder func(*http.Request, *Request) error

func HeaderClaim(header, claim string) RequestBuilder {
	return func(hr *http.Request, tr *Request) error {
		v := hr.Header.Get(header)
		if len(v) > 0 {
			tr.Claims[claim] = v
		}

		return nil
	}
}

func ParameterClaim(parameter, claim string) RequestBuilder {
	return func(hr *http.Request, tr *Request) error {
		v := hr.Form.Get(parameter)
		if len(v) > 0 {
			tr.Claims[claim] = v
			return nil
		}

		v = hr.PostForm.Get(parameter)
		if len(v) > 0 {
			tr.Claims[claim] = v
		}

		return nil
	}
}

func VariableClaim(variable, claim string) RequestBuilder {
	return func(hr *http.Request, tr *Request) error {
		v := mux.Vars(hr)[variable]
		if len(v) > 0 {
			tr.Claims[claim] = v
		}

		return nil
	}
}

func BuildRequest(hr *http.Request, b ...RequestBuilder) (*Request, error) {
	tr := &Request{
		Claims: make(map[string]interface{}),
	}

	for _, f := range b {
		if err := f(hr, tr); err != nil {
			return nil, err
		}
	}

	return tr, nil
}

func NewBuilders(d Descriptor) []RequestBuilder {
	var rbs []RequestBuilder
	for header, claim := range d.HeaderClaims {
		rbs = append(rbs, HeaderClaim(header, claim))
	}

	for parameter, claim := range d.ParameterClaims {
		rbs = append(rbs, ParameterClaim(parameter, claim))
	}

	return rbs
}

func DecodeRequest(b ...RequestBuilder) func(context.Context, *http.Request) (interface{}, error) {
	return func(_ context.Context, hr *http.Request) (interface{}, error) {
		if err := hr.ParseForm(); err != nil {
			return nil, err
		}

		tr, err := BuildRequest(hr, b...)
		if err != nil {
			return nil, err
		}

		return tr, nil
	}
}

func EncodeResponse(_ context.Context, response http.ResponseWriter, value interface{}) error {
	response.Header().Set("Content-Type", "application/jose")
	_, err := response.Write([]byte(value.(string)))
	return err
}
