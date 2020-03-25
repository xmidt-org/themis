package token

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"
)

func testNewRequestBuildersInvalidClaim(t *testing.T) {
	assert := assert.New(t)
	rb, err := NewRequestBuilders(Options{
		Claims: map[string]Value{
			"bad": Value{
				Header:    "xxx",
				Parameter: "yyy",
				Variable:  "zzz",
			},
		},
	})

	assert.Equal(ErrVariableNotAllowed, err)
	assert.Empty(rb)
}

func testNewRequestBuildersInvalidMetadata(t *testing.T) {
	assert := assert.New(t)
	rb, err := NewRequestBuilders(Options{
		Metadata: map[string]Value{
			"bad": Value{
				Header:    "xxx",
				Parameter: "yyy",
				Variable:  "zzz",
			},
		},
	})

	assert.Equal(ErrVariableNotAllowed, err)
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
				Claims: map[string]Value{
					"fromHeader": Value{
						Header: "X-Claim",
					},
					"missing": Value{
						Header: "X-Missing",
					},
				},
				Metadata: map[string]Value{
					"fromHeader": Value{
						Header: "X-Metadata",
					},
					"missing": Value{
						Header: "X-Missing",
					},
				},
				PartnerID: &PartnerID{
					Claim:    "partner-id-claim",
					Metadata: "partner-id-metadata",
					Header:   "X-Midt-Partner-ID",
				},
			},
			uri: "/test",
			header: http.Header{
				"X-Claim":           []string{"foo"},
				"X-Metadata":        []string{"bar"},
				"X-Midt-Partner-ID": []string{"test"},
			},
			expected: &Request{
				Claims: map[string]interface{}{
					"fromHeader":       "foo",
					"partner-id-claim": "test",
				},
				Metadata: map[string]interface{}{
					"fromHeader":          "bar",
					"partner-id-metadata": "test",
				},
			},
		},
		{
			options: Options{
				Claims: map[string]Value{
					"fromParameter": Value{
						Parameter: "claim",
					},
					"missing": Value{
						Parameter: "missing",
					},
				},
				Metadata: map[string]Value{
					"fromParameter": Value{
						Parameter: "metadata",
					},
					"missing": Value{
						Parameter: "missing",
					},
				},
				PartnerID: &PartnerID{
					Claim:     "partner-id-claim",
					Metadata:  "partner-id-metadata",
					Parameter: "pid",
				},
			},
			uri: "/test?pid=test&claim=foo&metadata=bar",
			expected: &Request{
				Claims: map[string]interface{}{
					"fromParameter":    "foo",
					"partner-id-claim": "test",
				},
				Metadata: map[string]interface{}{
					"fromParameter":       "bar",
					"partner-id-metadata": "test",
				},
			},
		},
		{
			options: Options{
				Claims: map[string]Value{
					"fromVariable": Value{
						Variable: "claim",
					},
				},
				Metadata: map[string]Value{
					"fromVariable": Value{
						Variable: "metadata",
					},
				},
				PartnerID: &PartnerID{
					Claim:     "partner-id-claim",
					Metadata:  "partner-id-metadata",
					Parameter: "pid",
					Default:   "test",
				},
			},
			uri: "/test/foo/bar",
			urlVariables: map[string]string{
				"claim":    "foo",
				"metadata": "bar",
			},
			expected: &Request{
				Claims: map[string]interface{}{
					"fromVariable":     "foo",
					"partner-id-claim": "test",
				},
				Metadata: map[string]interface{}{
					"fromVariable":        "bar",
					"partner-id-metadata": "test",
				},
			},
		},
		{
			options: Options{
				Claims: map[string]Value{
					"fromVariable": Value{
						Variable: "claim",
					},
				},
				Metadata: map[string]Value{
					"fromVariable": Value{
						Variable: "metadata",
					},
				},
			},
			uri: "/test/foo/bar",
			urlVariables: map[string]string{
				"claim":    "foo",
				"metadata": "bar",
			},
			expected: &Request{
				Claims: map[string]interface{}{
					"fromVariable": "foo",
				},
				Metadata: map[string]interface{}{
					"fromVariable": "bar",
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
			Claims: map[string]Value{
				"missing": Value{
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

			buildErr, ok := err.(BuildError)
			require.True(ok)
			assert.Equal(http.StatusBadRequest, buildErr.StatusCode())
		})
	}
}

func TestNewRequestBuilders(t *testing.T) {
	t.Run("InvalidClaim", testNewRequestBuildersInvalidClaim)
	t.Run("InvalidMetadata", testNewRequestBuildersInvalidMetadata)
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
				Claims:   map[string]interface{}{"claim": []int{1, 2, 3}},
				Metadata: make(map[string]interface{}),
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
				Claims:   make(map[string]interface{}),
				Metadata: map[string]interface{}{"metadata": -75.8},
			},
		},
		{
			builders: RequestBuilders{
				RequestBuilderFunc(func(_ *http.Request, r *Request) error {
					r.Claims["claim1"] = 238947123
					return nil
				}),
				RequestBuilderFunc(func(_ *http.Request, r *Request) error {
					r.Metadata["metadata1"] = "value1"
					return nil
				}),
				RequestBuilderFunc(func(_ *http.Request, r *Request) error {
					r.Claims["claim2"] = []byte{1, 2, 3}
					r.Metadata["metadata2"] = 15.7
					return nil
				}),
			},
			expected: &Request{
				Claims:   map[string]interface{}{"claim1": 238947123, "claim2": []byte{1, 2, 3}},
				Metadata: map[string]interface{}{"metadata1": "value1", "metadata2": 15.7},
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
				assert  = assert.New(t)
				require = require.New(t)

				request, actualErr = BuildRequest(httptest.NewRequest("GET", "/", nil), record)
			)

			buildErr, ok := actualErr.(BuildError)
			require.True(ok)

			assert.Contains(multierr.Errors(buildErr.Err), expectedErr)
			assert.Equal(http.StatusBadRequest, buildErr.StatusCode())
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
		(&DecodeClaimsError{}).Unwrap(),
	)

	assert.Equal(
		unwrappedErr,
		(&DecodeClaimsError{Err: unwrappedErr}).Unwrap(),
	)
}

func testDecodeClaimsErrorError(t *testing.T) {
	t.Run("NoNested", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			errText = (&DecodeClaimsError{
				URL:        "https://testy.com/foo?bar=1",
				StatusCode: 511,
			}).Error()
		)

		assert.Contains(errText, "https://testy.com/foo?bar=1")
		assert.Contains(errText, "511")
	})

	t.Run("WithNested", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			errText = (&DecodeClaimsError{
				URL:        "ftp://something.net",
				StatusCode: 499,
				Err:        errors.New("this is a nested error"),
			}).Error()
		)

		assert.Contains(errText, "ftp://something.net")
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
			err: &DecodeClaimsError{
				StatusCode: 475,
				URL:        "http://comcast.testy.test/moo",
			},
			expected: `{
				"url": "http://comcast.testy.test/moo",
				"statusCode": 475,
				"err": ""
			}`,
		},
		{
			err: &DecodeClaimsError{
				StatusCode: 314,
				URL:        "http://pi.numbers.com",
				Err:        errors.New("this is a nested error"),
			},
			expected: `{
				"url": "http://pi.numbers.com",
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
			Claims:   map[string]interface{}{"claim": "value"},
			Metadata: make(map[string]interface{}),
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

	buildErr, ok := actualErr.(BuildError)
	require.True(ok)

	assert.Contains(multierr.Errors(buildErr.Err), expectedErr)
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

	assert.Equal("application/jose", response.HeaderMap.Get("Content-Type"))
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
					Body:       ioutil.NopCloser(strings.NewReader(record.body)),
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
			Body:       ioutil.NopCloser(strings.NewReader("this is not JSON")),
			Request:    httptest.NewRequest("POST", "http://schmoogle.com", nil),
		}
	)

	v, err := DecodeRemoteClaimsResponse(context.Background(), response)
	assert.Nil(v)
	require.Error(err)
	require.IsType((*DecodeClaimsError)(nil), err)

	dce := err.(*DecodeClaimsError)
	assert.Equal(523, dce.StatusCode)
	assert.Equal("http://schmoogle.com", dce.URL)
	assert.Nil(dce.Err)
}

func testDecodeRemoteClaimsResponseBodyError(t *testing.T) {
	var (
		assert   = assert.New(t)
		require  = require.New(t)
		response = &http.Response{
			StatusCode: 466,
			Body: ioutil.NopCloser(
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
