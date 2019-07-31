package token

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"random/randomtest"
	"strconv"
	"testing"
	"time"
	"xhttp/xhttpclient"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestClaimBuilder(t *testing.T) {
	testData := []struct {
		request  *Request
		expected map[string]interface{}
	}{
		{
			request:  new(Request),
			expected: map[string]interface{}{},
		},
		{
			request: &Request{
				Claims: map[string]interface{}{"foo": 1, "bar": "value"},
			},
			expected: map[string]interface{}{"foo": 1, "bar": "value"},
		},
	}

	for i, record := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var (
				assert = assert.New(t)
				actual = make(map[string]interface{})
			)

			assert.NoError(
				requestClaimBuilder{}.AddClaims(context.Background(), record.request, actual),
			)

			assert.Equal(record.expected, actual)
		})
	}
}

func TestStaticClaimBuilder(t *testing.T) {
	testData := []struct {
		builder  staticClaimBuilder
		expected map[string]interface{}
	}{
		{
			expected: map[string]interface{}{},
		},
		{
			builder:  staticClaimBuilder{},
			expected: map[string]interface{}{},
		},
		{
			builder:  staticClaimBuilder{"foo": 1, "bar": "value"},
			expected: map[string]interface{}{"foo": 1, "bar": "value"},
		},
	}

	for i, record := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var (
				assert = assert.New(t)
				actual = make(map[string]interface{})
			)

			assert.NoError(
				record.builder.AddClaims(context.Background(), new(Request), actual),
			)

			assert.Equal(record.expected, actual)
		})
	}
}

func TestTimeClaimBuilder(t *testing.T) {
	var (
		expectedNow = time.Now()
		now         = func() time.Time { return expectedNow }
		testData    = []struct {
			builder  timeClaimBuilder
			expected map[string]interface{}
		}{
			{
				builder: timeClaimBuilder{
					now:              now,
					disableNotBefore: true,
				},
				expected: map[string]interface{}{
					"iat": expectedNow.UTC().Unix(),
				},
			},
			{
				builder: timeClaimBuilder{
					now: now,
				},
				expected: map[string]interface{}{
					"iat": expectedNow.UTC().Unix(),
					"nbf": expectedNow.UTC().Unix(),
				},
			},
			{
				builder: timeClaimBuilder{
					now:            now,
					duration:       24 * time.Hour,
					notBeforeDelta: 5 * time.Minute,
				},
				expected: map[string]interface{}{
					"iat": expectedNow.UTC().Unix(),
					"nbf": expectedNow.UTC().Add(5 * time.Minute).Unix(),
					"exp": expectedNow.UTC().Add(24 * time.Hour).Unix(),
				},
			},
			{
				builder: timeClaimBuilder{
					now:              now,
					duration:         30 * time.Minute,
					disableNotBefore: true,
				},
				expected: map[string]interface{}{
					"iat": expectedNow.UTC().Unix(),
					"exp": expectedNow.UTC().Add(30 * time.Minute).Unix(),
				},
			},
		}
	)

	for i, record := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var (
				assert = assert.New(t)
				actual = make(map[string]interface{})
			)

			assert.NoError(
				record.builder.AddClaims(context.Background(), new(Request), actual),
			)

			assert.Equal(record.expected, actual)
		})
	}
}

func TestNonceClaimBuilder(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var (
			assert = assert.New(t)
			noncer = new(randomtest.Noncer)

			actual  = make(map[string]interface{})
			builder = nonceClaimBuilder{n: noncer}
		)

		noncer.Expect("test", nil).Once()
		assert.NoError(builder.AddClaims(context.Background(), new(Request), actual))
		assert.Equal(
			map[string]interface{}{"jti": "test"},
			actual,
		)
	})

	t.Run("Error", func(t *testing.T) {
		var (
			assert      = assert.New(t)
			noncer      = new(randomtest.Noncer)
			expectedErr = errors.New("expected")

			actual  = make(map[string]interface{})
			builder = nonceClaimBuilder{n: noncer}
		)

		noncer.Expect("", expectedErr).Once()
		assert.Equal(
			expectedErr,
			builder.AddClaims(context.Background(), new(Request), actual),
		)

		assert.Empty(actual)
	})
}

func TestNewRemoteClaimBuilder(t *testing.T) {
	t.Run("NoURL", func(t *testing.T) {
		assert := assert.New(t)
		cb, err := newRemoteClaimBuilder(new(http.Client), nil, new(RemoteClaims))
		assert.Nil(cb)
		assert.Error(err)
	})

	t.Run("BadURL", func(t *testing.T) {
		var (
			assert       = assert.New(t)
			remoteClaims = &RemoteClaims{
				URL: "this is not valid (%$&@!()&*()*%",
			}
		)

		cb, err := newRemoteClaimBuilder(new(http.Client), nil, remoteClaims)
		assert.Nil(cb)
		assert.Error(err)
	})
}

func testRemoteClaimBuilderAddClaims(t *testing.T) {
	testData := []struct {
		method   string
		client   xhttpclient.Interface
		metadata map[string]interface{}
		request  *Request
		expected map[string]interface{}
	}{
		{
			request:  new(Request),
			expected: map[string]interface{}{"custom": "value"},
		},
		{
			request:  &Request{Metadata: map[string]interface{}{"request": "value"}},
			expected: map[string]interface{}{"request": "value", "custom": "value"},
		},
		{
			method:   http.MethodPut,
			client:   new(http.Client),
			metadata: map[string]interface{}{"external": "value"},
			request:  new(Request),
			expected: map[string]interface{}{"external": "value", "custom": "value"},
		},
		{
			method:   http.MethodPatch,
			client:   new(http.Client),
			metadata: map[string]interface{}{"external": "value"},
			request:  &Request{Metadata: map[string]interface{}{"request": "value"}},
			expected: map[string]interface{}{"external": "value", "request": "value", "custom": "value"},
		},
	}

	for i, record := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var (
				assert       = assert.New(t)
				require      = require.New(t)
				remoteClaims = &RemoteClaims{Method: record.method}

				handler = http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
					assert.Equal("application/json", request.Header.Get("Content-Type"))
					expectedMethod := record.method
					if len(expectedMethod) == 0 {
						expectedMethod = http.MethodPost
					}

					assert.Equal(expectedMethod, request.Method)

					b, err := ioutil.ReadAll(request.Body)
					assert.NoError(err)

					var input map[string]interface{}
					assert.NoError(json.Unmarshal(b, &input))
					if input == nil {
						input = make(map[string]interface{})
					}

					input["custom"] = "value"

					response.Header().Set("Content-Type", "application/json")
					b, err = json.Marshal(input)
					assert.NoError(err)
					assert.NotEmpty(b)
					response.Write(b)
				})
			)

			server := httptest.NewServer(handler)
			defer server.Close()
			remoteClaims.URL = server.URL

			builder, err := newRemoteClaimBuilder(
				record.client,
				record.metadata,
				remoteClaims,
			)

			require.NoError(err)
			require.NotNil(builder)

			actual := make(map[string]interface{})
			require.NoError(builder.AddClaims(context.Background(), record.request, actual))
			assert.Equal(record.expected, actual)
		})
	}
}

func testRemoteClaimBuilderRemoteError(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		handler = http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			response.WriteHeader(http.StatusInternalServerError)
		})
	)

	server := httptest.NewServer(handler)
	defer server.Close()

	builder, err := newRemoteClaimBuilder(nil, nil, &RemoteClaims{URL: server.URL})
	require.NoError(err)
	require.NotNil(builder)

	assert.Error(builder.AddClaims(context.Background(), new(Request), make(map[string]interface{})))
}

func TestRemoteClaimBuilder(t *testing.T) {
	t.Run("AddClaims", testRemoteClaimBuilderAddClaims)
	t.Run("RemoteError", testRemoteClaimBuilderRemoteError)
}

func testNewClaimBuildersMinimum(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		noncer = new(randomtest.Noncer)
	)

	builder, err := NewClaimBuilders(noncer, nil, Options{
		Nonce:       false,
		DisableTime: true,
	})

	require.NoError(err)
	require.NotEmpty(builder)

	actual := make(map[string]interface{})
	assert.NoError(
		builder.AddClaims(context.Background(), &Request{Claims: map[string]interface{}{"request": 123}}, actual),
	)

	assert.Equal(
		map[string]interface{}{"request": 123},
		actual,
	)

	noncer.AssertExpectations(t)
}

func testNewClaimBuildersBadValue(t *testing.T) {
	var (
		assert = assert.New(t)
		noncer = new(randomtest.Noncer)
	)

	builder, err := NewClaimBuilders(noncer, nil, Options{
		Nonce:       false,
		DisableTime: true,
		Claims: map[string]Value{
			"bad": Value{}, // the value should have something configured
		},
	})

	assert.Nil(builder)
	assert.Error(err)

	noncer.AssertExpectations(t)
}

func testNewClaimBuildersStatic(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		noncer = new(randomtest.Noncer)
	)

	builder, err := NewClaimBuilders(noncer, nil, Options{
		Nonce:       false,
		DisableTime: true,
		Claims: map[string]Value{
			"static1": Value{
				Value: -72.5,
			},
			"static2": Value{
				Value: []string{"a", "b"},
			},
			"http1": Value{
				Header: "X-Ignore-Me",
			},
		},
	})

	require.NoError(err)
	require.NotEmpty(builder)

	actual := make(map[string]interface{})
	assert.NoError(
		builder.AddClaims(context.Background(), &Request{Claims: map[string]interface{}{"request": 123}}, actual),
	)

	assert.Equal(
		map[string]interface{}{"static1": -72.5, "static2": []string{"a", "b"}, "request": 123},
		actual,
	)

	noncer.AssertExpectations(t)
}

func testNewClaimBuildersNoRemote(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		noncer = new(randomtest.Noncer)

		expectedNow = time.Now()
		now         = func() time.Time { return expectedNow }
	)

	builder, err := NewClaimBuilders(noncer, nil, Options{
		Nonce:          true,
		Duration:       24 * time.Hour,
		NotBeforeDelta: 15 * time.Second,
		Claims: map[string]Value{
			"static1": Value{
				Value: -72.5,
			},
			"static2": Value{
				Value: []string{"a", "b"},
			},
			"http1": Value{
				Header: "X-Ignore-Me",
			},
		},
	})

	require.NoError(err)
	require.NotEmpty(builder)

	for _, b := range builder {
		if tb, ok := b.(*timeClaimBuilder); ok {
			tb.now = now
			break
		}
	}

	noncer.Expect("test", nil).Once()

	actual := make(map[string]interface{})
	assert.NoError(
		builder.AddClaims(context.Background(), &Request{Claims: map[string]interface{}{"request": 123}}, actual),
	)

	assert.Equal(
		map[string]interface{}{
			"static1": -72.5,
			"static2": []string{"a", "b"},
			"request": 123,
			"jti":     "test",
			"iat":     expectedNow.UTC().Unix(),
			"nbf":     expectedNow.Add(15 * time.Second).UTC().Unix(),
			"exp":     expectedNow.Add(24 * time.Hour).UTC().Unix(),
		},
		actual,
	)

	noncer.AssertExpectations(t)
}

func testNewClaimBuildersFull(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		noncer = new(randomtest.Noncer)

		expectedNow = time.Now()
		now         = func() time.Time { return expectedNow }
		options     = Options{
			Nonce:          true,
			Duration:       24 * time.Hour,
			NotBeforeDelta: 15 * time.Second,
			Claims: map[string]Value{
				"static1": Value{
					Value: -72.5,
				},
				"static2": Value{
					Value: []string{"a", "b"},
				},
				"http1": Value{
					Header: "X-Ignore-Me",
				},
			},
			Metadata: map[string]Value{
				"extra1": Value{
					Value: "extra stuff",
				},
				"http2": Value{
					Parameter: "foo",
				},
			},
		}

		handler = http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			body, err := ioutil.ReadAll(request.Body)
			require.NoError(err)

			var metadata map[string]interface{}
			require.NoError(json.Unmarshal(body, &metadata))
			assert.Equal(map[string]interface{}{"extra1": "extra stuff"}, metadata)

			response.Header().Set("Content-Type", "application/json")
			response.Write([]byte(`{"remote": "value"}`))
		})
	)

	server := httptest.NewServer(handler)
	defer server.Close()
	options.Remote = &RemoteClaims{
		URL: server.URL,
	}

	builder, err := NewClaimBuilders(noncer, nil, options)
	require.NoError(err)
	require.NotEmpty(builder)

	for _, b := range builder {
		if tb, ok := b.(*timeClaimBuilder); ok {
			tb.now = now
			break
		}
	}

	noncer.Expect("test", nil).Once()

	actual := make(map[string]interface{})
	assert.NoError(
		builder.AddClaims(context.Background(), &Request{Claims: map[string]interface{}{"request": 123}}, actual),
	)

	assert.Equal(
		map[string]interface{}{
			"static1": -72.5,
			"static2": []string{"a", "b"},
			"request": 123,
			"remote":  "value",
			"jti":     "test",
			"iat":     expectedNow.UTC().Unix(),
			"nbf":     expectedNow.Add(15 * time.Second).UTC().Unix(),
			"exp":     expectedNow.Add(24 * time.Hour).UTC().Unix(),
		},
		actual,
	)

	noncer.AssertExpectations(t)
}

func TestNewClaimBuilders(t *testing.T) {
	t.Run("Minimal", testNewClaimBuildersMinimum)
	t.Run("BadValue", testNewClaimBuildersBadValue)
	t.Run("Static", testNewClaimBuildersStatic)
	t.Run("NoRemote", testNewClaimBuildersNoRemote)
	t.Run("Full", testNewClaimBuildersFull)
}
