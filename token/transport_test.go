// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package token

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xmidt-org/sallust"
	"go.uber.org/multierr"
)

func testNewRequestBuildersInvalidClaim(t *testing.T) {
	assert := assert.New(t)
	rb, err := NewRequestBuilders(Options{
		Claims: []Value{
			{
				// nolint:goconst
				Key: "bad",
				// nolint:goconst
				Header: "xxx",
				// nolint:goconst
				Parameter: "yyy",
				// nolint:goconst
				Variable: "zzz",
			},
		},
	})

	assert.ErrorIs(err, ErrVariableNotAllowed)
	assert.Empty(rb)
}

func testNewRequestBuildersInvalidMetadata(t *testing.T) {
	assert := assert.New(t)
	rb, err := NewRequestBuilders(Options{
		Metadata: []Value{
			{
				// nolint:goconst
				Key: "bad",
				// nolint:goconst
				Header: "xxx",
				// nolint:goconst
				Parameter: "yyy",
				// nolint:goconst
				Variable: "zzz",
			},
		},
	})

	assert.ErrorIs(err, ErrVariableNotAllowed)
	assert.Empty(rb)
}

func testNewRequestBuildersInvalidPathWildCards(t *testing.T) {
	assert := assert.New(t)
	rb, err := NewRequestBuilders(Options{
		PathWildCards: []Value{
			{
				// nolint:goconst
				Key: "bad",
				// nolint:goconst
				Header: "xxx",
				// nolint:goconst
				Parameter: "yyy",
				// nolint:goconst
				Variable: "zzz",
			},
		},
	})

	assert.ErrorIs(err, ErrVariableNotAllowed)
	assert.Empty(rb)
}

func testNewRequestBuildersInvalidQueryParameters(t *testing.T) {
	assert := assert.New(t)
	rb, err := NewRequestBuilders(Options{
		QueryParameters: []Value{
			{
				// nolint:goconst
				Key: "bad",
				// nolint:goconst
				Header: "xxx",
				// nolint:goconst
				Parameter: "yyy",
				// nolint:goconst
				Variable: "zzz",
			},
		},
	})

	assert.ErrorIs(err, ErrVariableNotAllowed)
	assert.Empty(rb)
}

func testNewRequestBuildersSuccess(t *testing.T) {
	testData := []struct {
		options      Options
		uri          string
		header       http.Header
		urlVariables map[string]string
		expected     *Request
	}{
		{
			uri:      "/test",
			expected: NewRequest(),
		},
		{
			options: Options{
				Claims: []Value{
					{
						// nolint:goconst
						Key: "fromHeader",
						// nolint:goconst
						Header: "X-Claim",
					},
					{
						// nolint:goconst
						Key: "missing",
						// nolint:goconst
						Header: "X-Missing",
					},
				},
				Metadata: []Value{
					{
						// nolint:goconst
						Key: "fromHeader",
						// nolint:goconst
						Header: "X-Metadata",
					},
					{
						// nolint:goconst
						Key: "missing",
						// nolint:goconst
						Header: "X-Missing",
					},
				},
				PathWildCards: []Value{
					{
						// nolint:goconst
						Key: "fromHeader",
						// nolint:goconst
						Header: "X-PathVlaue",
					},
				},
				QueryParameters: []Value{
					{
						// nolint:goconst
						Key: "fromHeader",
						// nolint:goconst
						Header: "Accept",
					},
				},
				PartnerID: &PartnerID{
					// nolint:goconst
					Claim: "partner-id-claim",
					// nolint:goconst
					Metadata: "partner-id-metadata",
					// nolint:goconst
					PathWildCard: "partner-id-pathWildCard",
					// nolint:goconst
					QueryParameter: "partner-id-queryParameter",
					// nolint:goconst
					Header: "X-Midt-Partner-ID",
				},
			},
			uri: "/test",
			header: http.Header{
				// nolint:goconst
				"X-Claim": []string{"foo"},
				// nolint:goconst
				"X-Metadata": []string{"bar"},
				// nolint:goconst
				"X-PathVlaue": []string{"foobar"},
				// nolint:goconst
				"X-Midt-Partner-ID": []string{"test"},
				// nolint:goconst
				"Accept": []string{"json"},
			},
			expected: &Request{
				Logger: sallust.Default(),
				Claims: map[string]interface{}{
					// nolint:goconst
					"fromHeader": "foo",
					// nolint:goconst
					"partner-id-claim": "test",
				},
				Metadata: map[string]interface{}{
					// nolint:goconst
					"fromHeader": "bar",
					// nolint:goconst
					"partner-id-metadata": "test",
				},
				PathWildCards: map[string]any{
					// nolint:goconst
					"fromHeader": "foobar",
					// nolint:goconst
					"partner-id-pathWildCard": "test",
				},
				QueryParameters: map[string]any{
					// nolint:goconst
					"fromHeader": "json",
					// nolint:goconst
					"partner-id-queryParameter": "test",
				},
			},
		},
		{
			options: Options{
				Claims: []Value{
					{
						// nolint:goconst
						Key: "fromParameter",
						// nolint:goconst
						Parameter: "claim",
					},
					{
						// nolint:goconst
						Key: "missing",
						// nolint:goconst
						Parameter: "missing",
					},
				},
				Metadata: []Value{
					{
						// nolint:goconst
						Key: "fromParameter",
						// nolint:goconst
						Parameter: "metadata",
					},
					{
						// nolint:goconst
						Key: "missing",
						// nolint:goconst
						Parameter: "missing",
					},
				},
				PathWildCards: []Value{
					{
						// nolint:goconst
						Key: "fromParameter",
						// nolint:goconst
						Parameter: "pathWildCard",
					},
					{
						// nolint:goconst
						Key: "missing",
						// nolint:goconst
						Parameter: "missing",
					},
				},
				QueryParameters: []Value{
					{
						// nolint:goconst
						Key: "fromParameter",
						// nolint:goconst
						Parameter: "queryParameter",
					},
				},
				PartnerID: &PartnerID{
					Claim:          "partner-id-claim",
					Metadata:       "partner-id-metadata",
					PathWildCard:   "partner-id-pathWildCard",
					QueryParameter: "partner-id-queryParameter",
					Parameter:      "pid",
				},
			},
			uri: "/test?pid=test&claim=foo&metadata=bar&pathWildCard=foobar&queryParameter=json",
			expected: &Request{
				Logger: sallust.Default(),
				Claims: map[string]interface{}{
					// nolint:goconst
					"fromParameter": "foo",
					// nolint:goconst
					"partner-id-claim": "test",
				},
				Metadata: map[string]interface{}{
					// nolint:goconst
					"fromParameter": "bar",
					// nolint:goconst
					"partner-id-metadata": "test",
				},
				PathWildCards: map[string]any{
					// nolint:goconst
					"fromParameter": "foobar",
					// nolint:goconst
					"partner-id-pathWildCard": "test",
				},
				QueryParameters: map[string]any{
					// nolint:goconst
					"fromParameter": "json",
					// nolint:goconst
					"partner-id-queryParameter": "test",
				},
			},
		},
		{
			options: Options{
				Claims: []Value{
					{
						// nolint:goconst
						Key: "fromVariable",
						// nolint:goconst
						Variable: "claim",
					},
				},
				Metadata: []Value{
					{
						// nolint:goconst
						Key: "fromVariable",
						// nolint:goconst
						Variable: "metadata",
					},
				},
				PathWildCards: []Value{
					{
						// nolint:goconst
						Key: "fromVariable",
						// nolint:goconst
						Variable: "pathWildCard",
					},
				},
				QueryParameters: []Value{
					{
						// nolint:goconst
						Key: "fromVariable",
						// nolint:goconst
						Variable: "queryParameter",
					},
				},
				PartnerID: &PartnerID{
					Claim:          "partner-id-claim",
					Metadata:       "partner-id-metadata",
					PathWildCard:   "partner-id-pathWildCard",
					QueryParameter: "partner-id-queryParameter",
					Parameter:      "pid",
					Default:        "test",
				},
			},
			uri: "/test/foo/bar/json",
			urlVariables: map[string]string{
				"claim":          "foo",
				"metadata":       "bar",
				"pathWildCard":   "foobar",
				"queryParameter": "json",
			},
			expected: &Request{
				Logger: sallust.Default(),
				Claims: map[string]interface{}{
					"fromVariable":     "foo",
					"partner-id-claim": "test",
				},
				Metadata: map[string]interface{}{
					"fromVariable":        "bar",
					"partner-id-metadata": "test",
				},
				PathWildCards: map[string]any{
					"fromVariable":            "foobar",
					"partner-id-pathWildCard": "test",
				},
				QueryParameters: map[string]any{
					"fromVariable":              "json",
					"partner-id-queryParameter": "test",
				},
			},
		},
		{
			options: Options{
				Claims: []Value{
					{
						Key:      "fromVariable",
						Variable: "claim",
					},
				},
				Metadata: []Value{
					{
						Key:      "fromVariable",
						Variable: "metadata",
					},
				},
				PathWildCards: []Value{
					{
						Key:      "fromVariable",
						Variable: "pathWildCard",
					},
				},
				QueryParameters: []Value{
					{
						Key:      "fromVariable",
						Variable: "queryParameter",
					},
					{
						Key:   "fromStatic",
						Value: "StaticValue0",
					},
				},
			},
			uri: "/test/foo/bar/foobar/json",
			urlVariables: map[string]string{
				"claim":          "foo",
				"metadata":       "bar",
				"pathWildCard":   "foobar",
				"queryParameter": "json",
			},
			expected: &Request{
				Logger: sallust.Default(),
				Claims: map[string]interface{}{
					"fromVariable": "foo",
				},
				Metadata: map[string]interface{}{
					"fromVariable": "bar",
				},
				PathWildCards: map[string]any{
					"fromVariable": "foobar",
				},
				QueryParameters: map[string]any{
					"fromStatic":   "StaticValue0",
					"fromVariable": "json",
				},
			},
		},
	}

	for i, record := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var (
				assert  = assert.New(t)
				require = require.New(t)

				rb, err = NewRequestBuilders(record.options)
			)

			require.NoError(err)

			actual := NewRequest()
			original := httptest.NewRequest("GET", record.uri, nil)
			for name, values := range record.header {
				for _, value := range values {
					original.Header.Add(name, value)
				}
			}

			require.NoError(original.ParseForm())
			original = mux.SetURLVars(original, record.urlVariables)

			assert.NoError(rb.Build(original, actual))
			assert.Equal(*record.expected, *actual)
		})
	}
}

func testNewRequestBuildersMissingVariable(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		options = Options{
			Claims: []Value{
				{
					Key:      "missing",
					Variable: "missing",
				},
			},
		}

		rb, err = NewRequestBuilders(options)
	)

	require.NoError(err)
	assert.Error(rb.Build(httptest.NewRequest("GET", "/test", nil), new(Request)))
}

func testNewRequestBuildersInvalidPartnerID(t *testing.T) {
	testData := []struct {
		invalidPartnerID string
	}{
		{"*"},
		{"*,,"},
		{",*,"},
		{"*,   "},
	}

	for i, record := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var (
				assert  = assert.New(t)
				require = require.New(t)

				tokenRequest = NewRequest()
				httpRequest  = httptest.NewRequest("GET", "/test", nil)

				rb, err = NewRequestBuilders(Options{
					PartnerID: &PartnerID{
						Header: "Test-Header",
						Claim:  "test-claim",
					},
				})
			)

			require.NoError(err)
			httpRequest.Header.Set("Test-Header", record.invalidPartnerID)

			err = rb.Build(httpRequest, tokenRequest)
			assert.Error(err)

			var buildErr BuildError
			assert.ErrorAs(err, &buildErr)
			assert.Equal(http.StatusBadRequest, buildErr.StatusCode())
		})
	}
}

func TestNewRequestBuilders(t *testing.T) {
	t.Run("InvalidClaim", testNewRequestBuildersInvalidClaim)
	t.Run("InvalidMetadata", testNewRequestBuildersInvalidMetadata)
	t.Run("InvalidPathWildCards", testNewRequestBuildersInvalidPathWildCards)
	t.Run("InvalidQueryParameters", testNewRequestBuildersInvalidQueryParameters)
	t.Run("MissingVariable", testNewRequestBuildersMissingVariable)
	t.Run("InvalidPartnerID", testNewRequestBuildersInvalidPartnerID)
	t.Run("Success", testNewRequestBuildersSuccess)
}

func testBuildRequestSuccess(t *testing.T) {
	testData := []struct {
		builders RequestBuilders
		expected *Request
	}{
		{
			expected: NewRequest(),
		},
		{
			builders: RequestBuilders{},
			expected: NewRequest(),
		},
		{
			builders: RequestBuilders{
				RequestBuilderFunc(func(_ *http.Request, r *Request) error {
					r.Claims["claim"] = []int{1, 2, 3}
					return nil
				}),
			},
			expected: &Request{
				Logger:          sallust.Default(),
				Claims:          map[string]interface{}{"claim": []int{1, 2, 3}},
				Metadata:        make(map[string]interface{}),
				PathWildCards:   make(map[string]interface{}),
				QueryParameters: make(map[string]any),
			},
		},
		{
			builders: RequestBuilders{
				RequestBuilderFunc(func(_ *http.Request, r *Request) error {
					r.Metadata["metadata"] = -75.8
					return nil
				}),
			},
			expected: &Request{
				Logger:          sallust.Default(),
				Claims:          make(map[string]interface{}),
				Metadata:        map[string]interface{}{"metadata": -75.8},
				PathWildCards:   make(map[string]interface{}),
				QueryParameters: make(map[string]any),
			},
		},
		{
			builders: RequestBuilders{
				RequestBuilderFunc(func(_ *http.Request, r *Request) error {
					// nolint:goconst
					r.Claims["claim1"] = 238947123
					return nil
				}),
				RequestBuilderFunc(func(_ *http.Request, r *Request) error {
					// nolint:goconst
					r.Metadata["metadata1"] = "value1"
					return nil
				}),
				RequestBuilderFunc(func(_ *http.Request, r *Request) error {
					// nolint:goconst
					r.Claims["claim2"] = []byte{1, 2, 3}
					// nolint:goconst
					r.Metadata["metadata2"] = 15.7
					return nil
				}),
			},
			expected: &Request{
				Logger:          sallust.Default(),
				Claims:          map[string]interface{}{"claim1": 238947123, "claim2": []byte{1, 2, 3}},
				Metadata:        map[string]interface{}{"metadata1": "value1", "metadata2": 15.7},
				PathWildCards:   make(map[string]interface{}),
				QueryParameters: make(map[string]any),
			},
		},
	}

	for i, record := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var (
				assert  = assert.New(t)
				require = require.New(t)

				actual, err = BuildRequest(httptest.NewRequest("GET", "/", nil), record.builders)
			)

			require.NoError(err)
			require.NotNil(actual)
			assert.Equal(*record.expected, *actual)
		})
	}
}

func testBuildRequestFailure(t *testing.T) {
	var (
		expectedErr = errors.New("expected")
		testData    = []RequestBuilders{
			{
				RequestBuilderFunc(func(_ *http.Request, r *Request) error {
					return expectedErr
				}),
			},
			{
				RequestBuilderFunc(func(_ *http.Request, r *Request) error {
					r.Claims["doesnotmatter"] = 1
					return nil
				}),
				RequestBuilderFunc(func(_ *http.Request, r *Request) error {
					return expectedErr
				}),
			},
			{
				RequestBuilderFunc(func(_ *http.Request, r *Request) error {
					return expectedErr
				}),
				RequestBuilderFunc(func(_ *http.Request, r *Request) error {
					r.Claims["doesnotmatter"] = 1
					return nil
				}),
			},
		}
	)

	for i, record := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var (
				assert = assert.New(t)
				//require = require.New(t)

				request, actualErr = BuildRequest(httptest.NewRequest("GET", "/", nil), record)
			)

			var be BuildError
			assert.ErrorAs(actualErr, &be)
			assert.Contains(multierr.Errors(be.Err), expectedErr)
			assert.Equal(http.StatusBadRequest, be.StatusCode())
			assert.Nil(request)
		})
	}
}

func TestBuildRequest(t *testing.T) {
	t.Run("Success", testBuildRequestSuccess)
	t.Run("Failure", testBuildRequestFailure)
}

func testDecodeClaimsErrorUnwrap(t *testing.T) {
	var (
		assert       = assert.New(t)
		unwrappedErr = errors.New("unwrapped")
	)

	assert.Nil(
		(&RemoteClaimsResponseError{}).Unwrap(),
	)

	assert.Equal(
		unwrappedErr,
		(&RemoteClaimsResponseError{Err: unwrappedErr}).Unwrap(),
	)
}

func testDecodeClaimsErrorError(t *testing.T) {
	t.Run("NoNested", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			errText = (&RemoteClaimsResponseError{
				StatusCode: 511,
			}).Error()
		)

		assert.Contains(errText, "511")
	})

	t.Run("WithNested", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			errText = (&RemoteClaimsResponseError{
				StatusCode: 499,
				Err:        errors.New("this is a nested error"),
			}).Error()
		)

		assert.Contains(errText, "499")
		assert.Contains(errText, "this is a nested error")
	})
}

func testDecodeClaimsErrorMarshalJSON(t *testing.T) {
	testData := []struct {
		err      error
		expected string
	}{
		{
			err: &RemoteClaimsResponseError{
				StatusCode: 475,
			},
			expected: `{
				"statusCode": 475,
				"err": ""
			}`,
		},
		{
			err: &RemoteClaimsResponseError{
				StatusCode: 314,
				Err:        errors.New("this is a nested error"),
			},
			expected: `{
				"statusCode": 314,
				"err": "this is a nested error"
			}`,
		},
	}

	for i, record := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var (
				assert  = assert.New(t)
				require = require.New(t)
			)

			actual, err := json.Marshal(record.err)
			require.NoError(err)
			assert.JSONEq(record.expected, string(actual))
		})
	}
}

func TestDecodeClaimsError(t *testing.T) {
	t.Run("Unwrap", testDecodeClaimsErrorUnwrap)
	t.Run("Error", testDecodeClaimsErrorError)
	t.Run("MarshalJSON", testDecodeClaimsErrorMarshalJSON)
}

func testDecodeServerRequestSuccess(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		builders = RequestBuilders{
			RequestBuilderFunc(func(_ *http.Request, r *Request) error {
				// nolint:goconst
				r.Claims["claim"] = "value"
				return nil
			}),
		}

		decoder = DecodeServerRequest(builders)
	)

	require.NotNil(decoder)
	v, err := decoder(context.Background(), httptest.NewRequest("GET", "/", nil))
	require.NoError(err)
	require.IsType((*Request)(nil), v)
	assert.Equal(
		Request{
			Logger:          sallust.Default(),
			Claims:          map[string]interface{}{"claim": "value"},
			Metadata:        make(map[string]interface{}),
			PathWildCards:   make(map[string]interface{}),
			QueryParameters: make(map[string]any),
		},
		*v.(*Request),
	)
}

func testDecodeServerRequestFailure(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		expectedErr = errors.New("expected")
		builders    = RequestBuilders{
			RequestBuilderFunc(func(_ *http.Request, r *Request) error {
				return expectedErr
			}),
		}

		decoder = DecodeServerRequest(builders)
	)

	require.NotNil(decoder)
	v, actualErr := decoder(context.Background(), httptest.NewRequest("GET", "/", nil))
	assert.Nil(v)

	var be BuildError
	assert.ErrorAs(actualErr, &be)

	assert.Contains(multierr.Errors(be.Err), expectedErr)
}

func TestDecodeServerRequest(t *testing.T) {
	t.Run("Success", testDecodeServerRequestSuccess)
	t.Run("Failure", testDecodeServerRequestFailure)
}

func TestEncodeIssueResponse(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		expectedValue = "expected"
		response      = httptest.NewRecorder()
	)

	require.NoError(
		EncodeIssueResponse(context.Background(), response, expectedValue),
	)

	assert.Equal("application/jose", response.Header().Get("Content-Type"))
	assert.Equal(expectedValue, response.Body.String())
}

func testDecodeRemoteClaimsResponseSuccess(t *testing.T) {
	testData := []struct {
		body     string
		expected map[string]interface{}
	}{
		{
			body: "",
		},
		{
			body:     "{}",
			expected: map[string]interface{}{},
		},
		{
			body:     `{"key1": "value1"}`,
			expected: map[string]interface{}{"key1": "value1"},
		},
	}

	for i, record := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var (
				assert  = assert.New(t)
				require = require.New(t)

				response = &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(record.body)),
				}
			)

			v, err := DecodeRemoteClaimsResponse(context.Background(), response)
			require.NoError(err)
			require.IsType(map[string]interface{}{}, v)
			assert.Equal(record.expected, v)
		})
	}
}

func testDecodeRemoteClaimsResponseFailure(t *testing.T) {
	var (
		assert   = assert.New(t)
		require  = require.New(t)
		response = &http.Response{
			StatusCode: 523,
			Body:       io.NopCloser(strings.NewReader("this is not JSON")),
			Request:    httptest.NewRequest("POST", "https://example.com", nil),
		}
	)

	v, err := DecodeRemoteClaimsResponse(context.Background(), response)
	assert.Nil(v)
	require.Error(err)
	require.IsType(RemoteClaimsResponseError{}, err)

	var dce RemoteClaimsResponseError
	assert.ErrorAs(err, &dce)
	assert.Equal(523, dce.StatusCode)
	assert.Contains(dce.Err.Error(), "this is not JSON")
}

func testDecodeRemoteClaimsResponseBodyError(t *testing.T) {
	var (
		assert   = assert.New(t)
		require  = require.New(t)
		response = &http.Response{
			StatusCode: 466,
			Body: io.NopCloser(
				iotest.TimeoutReader(strings.NewReader("gibberish")),
			),
			Request: httptest.NewRequest("POST", "http://cantreadbody.com", nil),
		}
	)

	v, err := DecodeRemoteClaimsResponse(context.Background(), response)
	assert.Nil(v)
	require.Error(err)
}

func TestDecodeRemoteClaimsResponse(t *testing.T) {
	t.Run("Success", testDecodeRemoteClaimsResponseSuccess)
	t.Run("Failure", testDecodeRemoteClaimsResponseFailure)
	t.Run("BodyError", testDecodeRemoteClaimsResponseBodyError)
}
