// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package xhttpserver

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testResponseHeadersThen(t *testing.T, responseHeaders ResponseHeaders) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		next = http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			response.WriteHeader(299)
		})

		request  = httptest.NewRequest("GET", "/", nil)
		response = httptest.NewRecorder()
	)

	decorated := responseHeaders.Then(next)
	require.NotNil(decorated)

	decorated.ServeHTTP(response, request)
	assert.Equal(299, response.Code)
	for name := range responseHeaders.Header {
		assert.Equal(
			responseHeaders.Header[name],
			response.Header()[http.CanonicalHeaderKey(name)],
		)
	}
}

func testResponseHeadersThenFunc(t *testing.T, responseHeaders ResponseHeaders) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		next = func(response http.ResponseWriter, request *http.Request) {
			response.WriteHeader(299)
		}

		request  = httptest.NewRequest("GET", "/", nil)
		response = httptest.NewRecorder()
	)

	decorated := responseHeaders.ThenFunc(next)
	require.NotNil(decorated)

	decorated.ServeHTTP(response, request)
	assert.Equal(299, response.Code)
	for name := range responseHeaders.Header {
		assert.Equal(
			responseHeaders.Header[name],
			response.Header()[http.CanonicalHeaderKey(name)],
		)
	}
}

func TestResponseHeaders(t *testing.T) {
	testData := []ResponseHeaders{
		ResponseHeaders{},
		ResponseHeaders{
			Header: http.Header{},
		},
		ResponseHeaders{Header: http.Header{
			"x-test": []string{"value"},
		}},
		ResponseHeaders{Header: http.Header{
			"X-Single": []string{"value"},
			"x-douBLe": []string{"value1", "value2"},
		}},
	}

	for i, responseHeaders := range testData {
		t.Run(fmt.Sprintf("Then/%d", i), func(t *testing.T) {
			testResponseHeadersThen(t, responseHeaders)
		})

		t.Run(fmt.Sprintf("ThenFunc/%d", i), func(t *testing.T) {
			testResponseHeadersThenFunc(t, responseHeaders)
		})
	}
}
