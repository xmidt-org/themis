package token

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-kit/kit/endpoint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIssueHandler(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		endpoint = endpoint.Endpoint(func(_ context.Context, v interface{}) (interface{}, error) {
			var output bytes.Buffer
			output.WriteString("endpoint=run")
			for key, value := range v.(*Request).Claims {
				output.WriteRune(',')
				fmt.Fprintf(&output, "%s=%s", key, value)
			}

			return output.String(), nil
		})

		builders = RequestBuilders{
			RequestBuilderFunc(func(original *http.Request, r *Request) error {
				r.Claims["claim"] = original.Header.Get("Claim")
				return nil
			}),
		}

		handler  = NewIssueHandler(endpoint, builders)
		response = httptest.NewRecorder()
		request  = httptest.NewRequest("POST", "/", nil)
	)

	require.NotNil(handler)
	request.Header.Set("Claim", "fromHeader")
	handler.ServeHTTP(response, request)
	assert.Equal("application/jose", response.HeaderMap.Get("Content-Type"))
	assert.Equal("endpoint=run,claim=fromHeader", response.Body.String())
}

func TestNewClaimsHandler(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		endpoint = endpoint.Endpoint(func(_ context.Context, v interface{}) (interface{}, error) {
			claims := map[string]interface{}{
				"endpoint": "run",
			}

			for key, value := range v.(*Request).Claims {
				claims[key] = value
			}

			return claims, nil
		})

		builders = RequestBuilders{
			RequestBuilderFunc(func(original *http.Request, r *Request) error {
				r.Claims["claim"] = original.Header.Get("Claim")
				return nil
			}),
		}

		handler  = NewClaimsHandler(endpoint, builders)
		response = httptest.NewRecorder()
		request  = httptest.NewRequest("POST", "/", nil)
	)

	require.NotNil(handler)
	request.Header.Set("Claim", "fromHeader")
	handler.ServeHTTP(response, request)
	assert.Regexp("application/json.*", response.HeaderMap.Get("Content-Type"))
	assert.JSONEq(
		`{"endpoint": "run", "claim": "fromHeader"}`,
		response.Body.String(),
	)
}
