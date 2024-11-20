// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
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

func TestNewIssueHandlerWithoutClaimHeaders(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		endpoint = endpoint.Endpoint(func(_ context.Context, v interface{}) (interface{}, error) {
			var (
				resp   = Response{}
				output bytes.Buffer
			)
			output.WriteString("endpoint=run")
			for key, value := range v.(*Request).Claims {
				output.WriteRune(',')
				fmt.Fprintf(&output, "%s=%s", key, value)
			}

			resp.Body = output.Bytes()

			return resp, nil
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
	assert.Equal("application/jose", response.Header().Get("Content-Type"))
	assert.Empty(response.Header().Get("claim"))
	assert.Equal("endpoint=run,claim=fromHeader", response.Body.String())
}

func TestNewIssueHandlerWithClaimHeaders(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		endpoint = endpoint.Endpoint(func(_ context.Context, v interface{}) (interface{}, error) {
			var (
				resp = Response{
					Claims:       make(map[string]interface{}),
					HeaderClaims: map[string]string{"claim": "HeaderClaim"},
				}
				output bytes.Buffer
			)
			output.WriteString("endpoint=run")
			for key, value := range v.(*Request).Claims {
				output.WriteRune(',')
				fmt.Fprintf(&output, "%s=%s", key, value)
				resp.Claims[key] = value
			}

			resp.Body = output.Bytes()

			return resp, nil
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
	assert.Equal("application/jose", response.Header().Get("Content-Type"))
	assert.Equal("fromHeader", response.Header().Get("HeaderClaim"))
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
	assert.Regexp("application/json.*", response.Header().Get("Content-Type"))
	assert.JSONEq(
		`{"endpoint": "run", "claim": "fromHeader"}`,
		response.Body.String(),
	)
}
