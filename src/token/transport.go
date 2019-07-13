package token

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

type TokenRequestBuilder func(*http.Request, *Request) error

func HeaderClaim(header, claim string) TokenRequestBuilder {
	return func(hr *http.Request, tr *Request) error {
		v := hr.Header.Get(header)
		if len(v) > 0 {
			tr.Claims[claim] = v
		}

		return nil
	}
}

func ParameterClaim(parameter, claim string) TokenRequestBuilder {
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

func VariableClaim(variable, claim string) TokenRequestBuilder {
	return func(hr *http.Request, tr *Request) error {
		v := mux.Vars(hr)[variable]
		if len(v) > 0 {
			tr.Claims[claim] = v
		}

		return nil
	}
}

func MetaHeader(header, key string) TokenRequestBuilder {
	return func(hr *http.Request, tr *Request) error {
		v := hr.Header.Get(header)
		if len(v) > 0 {
			tr.Meta[key] = v
		}

		return nil
	}
}

func MetaParameter(parameter, key string) TokenRequestBuilder {
	return func(hr *http.Request, tr *Request) error {
		v := hr.Form.Get(parameter)
		if len(v) > 0 {
			tr.Meta[key] = v
			return nil
		}

		v = hr.PostForm.Get(parameter)
		if len(v) > 0 {
			tr.Meta[key] = v
		}

		return nil
	}
}

func BuildRequest(hr *http.Request, b ...TokenRequestBuilder) (*Request, error) {
	tr := &Request{
		Claims: make(map[string]interface{}),
		Meta:   make(map[string]interface{}),
	}

	for _, f := range b {
		if err := f(hr, tr); err != nil {
			return nil, err
		}
	}

	return tr, nil
}

func NewTokenRequestBuilders(d Descriptor) []TokenRequestBuilder {
	var rbs []TokenRequestBuilder
	for header, claim := range d.HeaderClaims {
		rbs = append(rbs, HeaderClaim(header, claim))
	}

	for parameter, claim := range d.ParameterClaims {
		rbs = append(rbs, ParameterClaim(parameter, claim))
	}

	for header, key := range d.MetaHeaders {
		rbs = append(rbs, MetaHeader(header, key))
	}

	for parameter, key := range d.MetaParameters {
		rbs = append(rbs, MetaParameter(parameter, key))
	}

	return rbs
}

func DecodeServerRequest(b ...TokenRequestBuilder) func(context.Context, *http.Request) (interface{}, error) {
	return func(_ context.Context, hr *http.Request) (interface{}, error) {
		tr, err := BuildRequest(hr, b...)
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
