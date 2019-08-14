package xhttpserver

import (
	"bytes"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"xlog"
	"xlog/xloghttp"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testNewServerChainNone(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		output bytes.Buffer
		base   = log.NewJSONLogger(&output)

		next = http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			_, ok := response.(TrackingWriter)
			assert.False(ok)
			assert.Equal(xlog.Default(), xlog.Get(request.Context()))

			response.WriteHeader(299)
		})

		chain = NewServerChain(
			Options{
				DisableTracking:      true,
				DisableHandlerLogger: true,
			},
			base,
		)

		response = httptest.NewRecorder()
		request  = httptest.NewRequest("POST", "/foo", nil)
	)

	decorated := chain.Then(next)
	require.NotNil(decorated)
	decorated.ServeHTTP(response, request)
	assert.Equal(299, response.Code)
}

func testNewServerChainHeaders(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		output bytes.Buffer
		base   = log.NewJSONLogger(&output)

		next = http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			_, ok := response.(TrackingWriter)
			assert.False(ok)
			assert.Equal(xlog.Default(), xlog.Get(request.Context()))
			assert.Equal("value", response.Header().Get("X-From-Configuration"))

			response.WriteHeader(299)
		})

		chain = NewServerChain(
			Options{
				Header: http.Header{
					"X-From-Configuration": []string{"value"},
				},
				DisableTracking:      true,
				DisableHandlerLogger: true,
			},
			base,
		)

		response = httptest.NewRecorder()
		request  = httptest.NewRequest("POST", "/foo", nil)
	)

	decorated := chain.Then(next)
	require.NotNil(decorated)
	decorated.ServeHTTP(response, request)
	assert.Equal(299, response.Code)
}

func testNewServerChainTracking(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		output bytes.Buffer
		base   = log.NewJSONLogger(&output)

		next = http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			assert.Implements((*TrackingWriter)(nil), response)
			assert.Equal(xlog.Default(), xlog.Get(request.Context()))
			assert.Equal("value", response.Header().Get("X-From-Configuration"))

			response.WriteHeader(299)
		})

		chain = NewServerChain(
			Options{
				Header: http.Header{
					"X-From-Configuration": []string{"value"},
				},
				DisableHandlerLogger: true,
			},
			base,
		)

		response = httptest.NewRecorder()
		request  = httptest.NewRequest("POST", "/foo", nil)
	)

	decorated := chain.Then(next)
	require.NotNil(decorated)
	decorated.ServeHTTP(response, request)
	assert.Equal(299, response.Code)
}

func testNewServerChainFull(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		output bytes.Buffer
		base   = log.NewJSONLogger(&output)

		next = http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			assert.Implements((*TrackingWriter)(nil), response)
			assert.Equal("value", response.Header().Get("X-From-Configuration"))
			xlog.Get(request.Context()).Log("foo", "bar")

			response.WriteHeader(299)
		})

		chain = NewServerChain(
			Options{
				Header: http.Header{
					"X-From-Configuration": []string{"value"},
				},
			},
			base,
			xloghttp.Method("requestMethod"),
			xloghttp.URI("requestURI"),
		)

		response = httptest.NewRecorder()
		request  = httptest.NewRequest("POST", "/foo", nil)
	)

	decorated := chain.Then(next)
	require.NotNil(decorated)
	decorated.ServeHTTP(response, request)
	assert.Equal(299, response.Code)
	assert.Contains(output.String(), "requestMethod")
	assert.Contains(output.String(), "POST")
	assert.Contains(output.String(), "requestURI")
	assert.Contains(output.String(), "/foo")
}

func TestNewServerChain(t *testing.T) {
	t.Run("None", testNewServerChainNone)
	t.Run("Headers", testNewServerChainHeaders)
	t.Run("Tracking", testNewServerChainTracking)
	t.Run("Full", testNewServerChainFull)
}

func testNewSimple(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		output bytes.Buffer
		base   = log.NewJSONLogger(&output)
		router = mux.NewRouter()

		s = New(
			Options{
				Address:               ":8080",
				MaxHeaderBytes:        1111,
				IdleTimeout:           19 * time.Minute,
				ReadHeaderTimeout:     27 * time.Second,
				ReadTimeout:           1239 * time.Hour,
				WriteTimeout:          289 * time.Millisecond,
				LogConnectionState:    false,
				DisableHTTPKeepAlives: true,
			},
			base,
			router,
		)
	)

	// give the router some state for reliable equals testing
	router.HandleFunc("/", func(http.ResponseWriter, *http.Request) {})

	require.NotNil(s)
	require.IsType((*http.Server)(nil), s)
	assert.Equal(":8080", s.(*http.Server).Addr)
	assert.Equal(router, s.(*http.Server).Handler)
	assert.Equal(1111, s.(*http.Server).MaxHeaderBytes)
	assert.Equal(19*time.Minute, s.(*http.Server).IdleTimeout)
	assert.Equal(27*time.Second, s.(*http.Server).ReadHeaderTimeout)
	assert.Equal(1239*time.Hour, s.(*http.Server).ReadTimeout)
	assert.Equal(289*time.Millisecond, s.(*http.Server).WriteTimeout)

	require.NotNil(s.(*http.Server).ErrorLog)
	s.(*http.Server).ErrorLog.Print("foo", "bar")
	assert.Greater(output.Len(), 0)

	assert.Nil(s.(*http.Server).ConnState)
}

func testNewFull(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		output bytes.Buffer
		base   = log.NewJSONLogger(&output)
		router = mux.NewRouter()

		s = New(
			Options{
				Address:            ":12000",
				MaxHeaderBytes:     192854,
				IdleTimeout:        9 * time.Hour,
				ReadHeaderTimeout:  113 * time.Minute,
				ReadTimeout:        99 * time.Second,
				WriteTimeout:       8456 * time.Nanosecond,
				LogConnectionState: true,
			},
			base,
			router,
		)
	)

	// give the router some state for reliable equals testing
	router.HandleFunc("/", func(http.ResponseWriter, *http.Request) {})

	require.NotNil(s)
	require.IsType((*http.Server)(nil), s)
	assert.Equal(":12000", s.(*http.Server).Addr)
	assert.Equal(router, s.(*http.Server).Handler)
	assert.Equal(192854, s.(*http.Server).MaxHeaderBytes)
	assert.Equal(9*time.Hour, s.(*http.Server).IdleTimeout)
	assert.Equal(113*time.Minute, s.(*http.Server).ReadHeaderTimeout)
	assert.Equal(99*time.Second, s.(*http.Server).ReadTimeout)
	assert.Equal(8456*time.Nanosecond, s.(*http.Server).WriteTimeout)

	require.NotNil(s.(*http.Server).ErrorLog)
	s.(*http.Server).ErrorLog.Print("foo", "bar")
	assert.Greater(output.Len(), 0)

	require.NotNil(s.(*http.Server).ConnState)
	output.Reset()
	s.(*http.Server).ConnState(new(net.IPConn), http.StateNew)
	assert.Greater(output.Len(), 0)
}

func TestNew(t *testing.T) {
	t.Run("Simple", testNewSimple)
	t.Run("Full", testNewFull)
}
