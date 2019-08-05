package token

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"xhttp/xhttpserver"

	"github.com/gorilla/mux"
)

var (
	ErrVariableNotAllowed = errors.New("Either header/parameter or variable can specified, but not all three")
)

// RequestBuilder is a strategy for building a token factory Request from an HTTP request
type RequestBuilder interface {
	Build(*http.Request, *Request) error
}

type RequestBuilderFunc func(*http.Request, *Request) error

func (rbf RequestBuilderFunc) Build(original *http.Request, tr *Request) error {
	return rbf(original, tr)
}

// RequestBuilders represents a set of RequestBuilder strategies that can be invoked in sequence
type RequestBuilders []RequestBuilder

func (rbs RequestBuilders) Build(original *http.Request, tr *Request) error {
	for _, rb := range rbs {
		if err := rb.Build(original, tr); err != nil {
			return err
		}
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
	required  bool
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

	if hprb.required {
		return xhttpserver.MissingValueError{
			Header:    hprb.header,
			Parameter: hprb.parameter,
		}
	}

	return nil
}

type variableRequestBuilder struct {
	key      string
	variable string
	required bool
	setter   func(string, interface{}, *Request)
}

func (vrb variableRequestBuilder) Build(original *http.Request, tr *Request) error {
	value := mux.Vars(original)[vrb.variable]
	if len(value) > 0 {
		vrb.setter(vrb.key, value, tr)
		return nil
	}

	if vrb.required {
		return xhttpserver.MissingVariableError{Variable: vrb.variable}
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
					required:  value.Required,
					setter:    claimsSetter,
				},
			)
		} else if len(value.Variable) > 0 {
			rb = append(rb,
				variableRequestBuilder{
					key:      name,
					variable: value.Variable,
					required: value.Required,
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
					required:  value.Required,
					setter:    metadataSetter,
				},
			)
		} else if len(value.Variable) > 0 {
			rb = append(rb,
				variableRequestBuilder{
					key:      name,
					variable: value.Variable,
					required: value.Required,
					setter:   metadataSetter,
				},
			)
		}
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
	return func(_ context.Context, hr *http.Request) (interface{}, error) {
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

func DecodeRemoteClaimsResponse(_ context.Context, response *http.Response) (interface{}, error) {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
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
