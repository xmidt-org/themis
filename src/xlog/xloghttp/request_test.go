package xloghttp

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"xlog"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParameters(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		var (
			assert = assert.New(t)

			output   bytes.Buffer
			original = log.NewJSONLogger(&output)

			p Parameters
		)

		assert.Equal(original, p.Use(original))
	})

	t.Run("NonEmpty", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)

			output   bytes.Buffer
			original = log.NewJSONLogger(&output)

			p Parameters
		)

		p.Add("key", "value")
		logger := p.Use(original)
		require.NotEqual(original, logger)

		logger.Log("msg", "hi")
		assert.Contains(output.String(), "key")
		assert.Contains(output.String(), "value")
	})
}

func TestMethod(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		request = httptest.NewRequest("GET", "/test?foo=bar", nil)
		p       Parameters
		builder = Method("requestMethod")
	)

	require.NotNil(builder)
	builder(request, &p)
	assert.Equal([]interface{}{"requestMethod", "GET"}, p.values)
}

func TestURI(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		request = httptest.NewRequest("GET", "/test?foo=bar", nil)
		p       Parameters
		builder = URI("requestURI")
	)

	require.NotNil(builder)
	builder(request, &p)
	assert.Equal([]interface{}{"requestURI", "/test"}, p.values)
}

func TestRemoteAddress(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		request = httptest.NewRequest("GET", "/test?foo=bar", nil)
		p       Parameters
		builder = RemoteAddress("remoteAddress")
	)

	require.NotNil(builder)
	request.RemoteAddr = "foobar.net"
	builder(request, &p)
	assert.Equal([]interface{}{"remoteAddress", "foobar.net"}, p.values)
}

func TestHeader(t *testing.T) {
	t.Run("NoValue", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)

			request = httptest.NewRequest("GET", "/test?foo=bar", nil)
			p       Parameters
			builder = Header("X-Test")
		)

		require.NotNil(builder)
		builder(request, &p)
		assert.Equal([]interface{}{"X-Test", ""}, p.values)
	})

	t.Run("SingleValue", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)

			request = httptest.NewRequest("GET", "/test?foo=bar", nil)
			p       Parameters
			builder = Header("X-Test")
		)

		require.NotNil(builder)
		request.Header.Set("X-Test", "value")
		builder(request, &p)
		assert.Equal([]interface{}{"X-Test", "value"}, p.values)
	})

	t.Run("NonCanonical", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)

			request = httptest.NewRequest("GET", "/test?foo=bar", nil)
			p       Parameters
			builder = Header("x-test")
		)

		require.NotNil(builder)
		request.Header.Set("x-test", "value")
		builder(request, &p)
		assert.Equal([]interface{}{"X-Test", "value"}, p.values)
	})

	t.Run("MultiValue", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)

			request = httptest.NewRequest("GET", "/test?foo=bar", nil)
			p       Parameters
			builder = Header("X-Test")
		)

		require.NotNil(builder)
		request.Header.Add("X-Test", "value1")
		request.Header.Add("X-Test", "value2")
		builder(request, &p)
		assert.Equal([]interface{}{"X-Test", "value1,value2"}, p.values)
	})
}

func TestParameter(t *testing.T) {
	t.Run("NoValue", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)

			request = httptest.NewRequest("GET", "/test?foo=value", nil)
			p       Parameters
			builder = Parameter("name")
		)

		require.NotNil(builder)
		require.NoError(request.ParseForm())
		builder(request, &p)
		assert.Equal([]interface{}{"name", ""}, p.values)
	})

	t.Run("SingleValue", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)

			request = httptest.NewRequest("GET", "/test?name=value", nil)
			p       Parameters
			builder = Parameter("name")
		)

		require.NotNil(builder)
		require.NoError(request.ParseForm())
		builder(request, &p)
		assert.Equal([]interface{}{"name", "value"}, p.values)
	})

	t.Run("MultiValue", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)

			request = httptest.NewRequest("GET", "/test?name=value1&name=value2", nil)
			p       Parameters
			builder = Parameter("name")
		)

		require.NotNil(builder)
		require.NoError(request.ParseForm())
		builder(request, &p)
		assert.Equal([]interface{}{"name", "value1,value2"}, p.values)
	})
}

func TestVariable(t *testing.T) {
	t.Run("NoValue", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)

			request = httptest.NewRequest("GET", "/test?foo=bar", nil)
			p       Parameters
			builder = Variable("name")
		)

		require.NotNil(builder)
		builder(request, &p)
		assert.Equal([]interface{}{"name", ""}, p.values)
	})

	t.Run("SingleValue", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)

			request = httptest.NewRequest("GET", "/test?foo=bar", nil)
			p       Parameters
			builder = Variable("name")
		)

		require.NotNil(builder)
		request = mux.SetURLVars(request, map[string]string{"name": "value"})
		builder(request, &p)
		assert.Equal([]interface{}{"name", "value"}, p.values)
	})
}

func TestWithRequest(t *testing.T) {
	t.Run("NoBuilders", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)

			output   bytes.Buffer
			original = log.NewJSONLogger(&output)

			request = WithRequest(httptest.NewRequest("GET", "/foo/bar", nil), original)
		)

		require.NotNil(request)
		assert.Equal(original, xlog.Get(request.Context()))
	})

	t.Run("WithBuilders", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)

			output   bytes.Buffer
			original = log.NewJSONLogger(&output)

			request = WithRequest(
				httptest.NewRequest("GET", "/foo/bar", nil),
				original,
				Method("requestMethod"),
				URI("requestURI"),
			)
		)

		require.NotNil(request)

		logger := xlog.Get(request.Context())
		require.NotEqual(original, logger)

		logger.Log("msg", "hi")
		assert.Contains(output.String(), "requestMethod")
		assert.Contains(output.String(), request.Method)
		assert.Contains(output.String(), "requestURI")
		assert.Contains(output.String(), request.URL.Path)
	})
}

func TestLogging(t *testing.T) {
	t.Run("NoBuilders", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)

			output   bytes.Buffer
			original = log.NewJSONLogger(&output)

			delegateCalled = false
			delegate       = http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
				delegateCalled = true
				logger := xlog.Get(request.Context())
				assert.Equal(original, logger)
			})

			logging = Logging{Base: original}
		)

		decorated := logging.Then(delegate)
		require.NotNil(decorated)
		decorated.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
		assert.True(delegateCalled)
	})

	t.Run("WithBuilders", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)

			output   bytes.Buffer
			original = log.NewJSONLogger(&output)

			delegateCalled = false
			delegate       = http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
				delegateCalled = true
				logger := xlog.Get(request.Context())
				require.NotEqual(original, logger)
				logger.Log("msg", "hi")
			})

			logging = Logging{
				Base:     original,
				Builders: []ParameterBuilder{Method("requestMethod")},
			}
		)

		decorated := logging.Then(delegate)
		require.NotNil(decorated)
		decorated.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
		assert.True(delegateCalled)

		assert.Contains(output.String(), "requestMethod")
		assert.Contains(output.String(), "GET")
	})
}
