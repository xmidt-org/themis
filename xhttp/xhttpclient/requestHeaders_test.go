package xhttpclient

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequestHeadersThen(t *testing.T, requestHeaders RequestHeaders) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		request          = httptest.NewRequest("GET", "/", nil)
		expectedResponse = new(http.Response)
		expectedErr      = errors.New("expected")

		roundTripper = new(mockRoundTripper)
	)

	decorated := requestHeaders.Then(roundTripper)
	require.NotNil(decorated)

	roundTripper.ExpectRoundTrip(request).Once().Return(expectedResponse, expectedErr)
	actualResponse, actualErr := decorated.RoundTrip(request) //nolint: bodyclose
	assert.Equal(expectedResponse, actualResponse)            //nolint: bodyclose
	assert.Equal(expectedErr, actualErr)

	for name := range requestHeaders.Header {
		assert.Equal(
			requestHeaders.Header[name],
			request.Header[http.CanonicalHeaderKey(name)],
		)
	}

	roundTripper.AssertExpectations(t)
}

func testRequestHeadersThenFunc(t *testing.T, requestHeaders RequestHeaders) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		request          = httptest.NewRequest("GET", "/", nil)
		expectedResponse = new(http.Response)
		expectedErr      = errors.New("expected")

		roundTripper = new(mockRoundTripper)
	)

	decorated := requestHeaders.ThenFunc(roundTripper.RoundTrip)
	require.NotNil(decorated)

	roundTripper.ExpectRoundTrip(request).Once().Return(expectedResponse, expectedErr)
	actualResponse, actualErr := decorated.RoundTrip(request) //nolint:bodyclose
	assert.Equal(expectedResponse, actualResponse)
	assert.Equal(expectedErr, actualErr)

	for name := range requestHeaders.Header {
		assert.Equal(
			requestHeaders.Header[name],
			request.Header[http.CanonicalHeaderKey(name)],
		)
	}

	roundTripper.AssertExpectations(t)
}

func TestRequestHeaders(t *testing.T) {
	testData := []RequestHeaders{
		RequestHeaders{},
		RequestHeaders{
			Header: http.Header{},
		},
		RequestHeaders{Header: http.Header{
			"x-test": []string{"value"},
		}},
		RequestHeaders{Header: http.Header{
			"X-Single": []string{"value"},
			"x-douBLe": []string{"value1", "value2"},
		}},
	}

	for i, requestHeaders := range testData {
		t.Run(fmt.Sprintf("Then/%d", i), func(t *testing.T) {
			testRequestHeadersThen(t, requestHeaders)
		})

		t.Run(fmt.Sprintf("ThenFunc/%d", i), func(t *testing.T) {
			testRequestHeadersThenFunc(t, requestHeaders)
		})
	}
}
