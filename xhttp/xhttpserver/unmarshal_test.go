package xhttpserver

import (
	"errors"
	"net/http"
	"testing"

	"github.com/xmidt-org/candlelight"
	"github.com/xmidt-org/sallust"
	"github.com/xmidt-org/themis/config"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestServerNotConfiguredError(t *testing.T) {
	var (
		assert = assert.New(t)

		err error = ServerNotConfiguredError{Key: "serverKey"}
	)

	assert.Contains(err.Error(), "serverKey")
}

func testUnmarshalProvideFull(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		handler = func(response http.ResponseWriter, request *http.Request) {
			assert.Equal("global value", response.Header().Get("X-Global"))
			assert.Equal("server", response.Header().Get("X-Factory"))
			assert.Equal("adhoc value", response.Header().Get("X-Adhoc"))
			response.WriteHeader(299)
		}

		router *mux.Router
		app    = fxtest.New(t,
			fx.Provide(
				sallust.Default,
				config.ProvideViper(
					config.Json(`
						{
							"server": {
								"address": "127.0.0.1:0",
								"disableHTTPKeepAlives": true
							}
						}
					`),
				),
				func() ChainFactory {
					return ChainFactoryFunc(func(name string, o Options) (alice.Chain, error) {
						return alice.New(
							ResponseHeaders{
								Header: http.Header{"X-Factory": []string{name}},
							}.Then,
						), nil
					})
				},
				Unmarshal{
					Key: "server",
					Chain: alice.New(ResponseHeaders{
						Header: http.Header{"X-Adhoc": []string{"adhoc value"}},
					}.Then),
				}.Provide,
			),
			fx.Populate(&router),
		)
	)

	require.NotNil(router)
	router.HandleFunc("/test", handler)
	app.RequireStart()
	app.RequireStop()
}

type testUnmarshalProvideOptionalIn struct {
	fx.In

	Router *mux.Router
}

func testUnmarshalProvideOptional(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		router *mux.Router
		app    = fxtest.New(t,
			fx.Provide(
				sallust.Default,
				config.ProvideViper(),
				Unmarshal{
					Key:      "server",
					Optional: true,
					Chain: alice.New(ResponseHeaders{
						Header: http.Header{"X-Adhoc": []string{"adhoc value"}},
					}.Then),
				}.Provide,
			),
			fx.Invoke(
				func(in testUnmarshalProvideOptionalIn) {
					assert.Nil(in.Router)
					router = in.Router
				},
			),
		)
	)

	require.Nil(router)
	app.RequireStart()
	app.RequireStop()
}

func testUnmarshalProvideRequired(t *testing.T) {
	var (
		assert = assert.New(t)

		app = fx.New(
			fx.Provide(
				fx.Logger(sallust.Printer{}),
				sallust.Default,
				config.ProvideViper(),
				Unmarshal{Key: "server"}.Provide,
			),
			fx.Invoke(
				func(*mux.Router) {
					assert.Fail("This invoke function should not have been called")
				},
			),
		)
	)

	assert.Error(app.Err())
}

func testUnmarshalProvideUnmarshalError(t *testing.T) {
	var (
		assert = assert.New(t)

		app = fx.New(
			fx.Provide(
				fx.Logger(sallust.Printer{}),
				sallust.Default,
				config.ProvideViper(
					config.Json(`
						{
							"server": {
								"disableHTTPKeepAlives": "this is not a bool"
							}
						}
					`),
				),
				Unmarshal{Key: "server"}.Provide,
			),
			fx.Invoke(
				func(*mux.Router) {
					assert.Fail("This invoke function should not have been called")
				},
			),
		)
	)

	assert.Error(app.Err())
}

func testUnmarshalProvideChainFactoryError(t *testing.T) {
	var (
		assert      = assert.New(t)
		expectedErr = errors.New("expected chain factory error")

		app = fx.New(
			fx.Provide(
				fx.Logger(sallust.Printer{}),
				sallust.Default,
				config.ProvideViper(
					config.Json(`
						{
							"server": {
								"address": "127.0.0.1:0",
								"disableHTTPKeepAlives": true
							}
						}
					`),
				),
				func() ChainFactory {
					return ChainFactoryFunc(func(name string, o Options) (alice.Chain, error) {
						return alice.Chain{}, expectedErr
					})
				},
				Unmarshal{Key: "server"}.Provide,
			),
			fx.Invoke(
				func(*mux.Router) {
					assert.Fail("This invoke function should not have been called")
				},
			),
		)
	)

	assert.Error(app.Err())
}

type testUnmarshalAnnotatedFullIn struct {
	fx.In

	Router *mux.Router `name:"server" optional:"true"`
}

func testUnmarshalAnnotatedFull(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		handler = func(response http.ResponseWriter, request *http.Request) {
			assert.Equal("global value", response.Header().Get("X-Global"))
			assert.Equal("server", response.Header().Get("X-Factory"))
			assert.Equal("adhoc value", response.Header().Get("X-Adhoc"))
			response.WriteHeader(299)
		}

		router *mux.Router
		app    = fxtest.New(t,
			fx.Provide(
				sallust.Default,
				config.ProvideViper(
					config.Json(`
						{
							"server": {
								"address": "127.0.0.1:0",
								"disableHTTPKeepAlives": true
							}
						}
					`),
				),
				func() ChainFactory {
					return ChainFactoryFunc(func(name string, o Options) (alice.Chain, error) {
						return alice.New(
							ResponseHeaders{
								Header: http.Header{"X-Factory": []string{name}},
							}.Then,
						), nil
					})
				},
				Unmarshal{
					Key: "server",
					Chain: alice.New(ResponseHeaders{
						Header: http.Header{"X-Adhoc": []string{"adhoc value"}},
					}.Then),
				}.Annotated(),
				func() (candlelight.Tracing, error) {
					return candlelight.New(candlelight.Config{Provider: "stdout"})
				},
			),
			fx.Invoke(
				func(in testUnmarshalAnnotatedFullIn) {
					assert.NotNil(in.Router)
					router = in.Router
				},
			),
		)
	)

	require.NotNil(router)
	router.HandleFunc("/test", handler)
	app.RequireStart()
	app.RequireStop()
}

type testUnmarshalAnnotatedNamedIn struct {
	fx.In

	Router *mux.Router `name:"componentName" optional:"true"`
}

func testUnmarshalAnnotatedNamed(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		handler = func(response http.ResponseWriter, request *http.Request) {
			assert.Equal("global value", response.Header().Get("X-Global"))
			assert.Equal("server", response.Header().Get("X-Factory"))
			assert.Equal("adhoc value", response.Header().Get("X-Adhoc"))
			response.WriteHeader(299)
		}

		router *mux.Router
		app    = fxtest.New(t,
			fx.Provide(
				sallust.Default,
				config.ProvideViper(
					config.Json(`
						{
							"server": {
								"address": "127.0.0.1:0",
								"disableHTTPKeepAlives": true
							}
						}
					`),
				),
				Unmarshal{
					Key:  "server",
					Name: "componentName",
					Chain: alice.New(ResponseHeaders{
						Header: http.Header{"X-Adhoc": []string{"adhoc value"}},
					}.Then),
				}.Annotated(),
			),
			fx.Invoke(
				func(in testUnmarshalAnnotatedNamedIn) {
					assert.NotNil(in.Router)
					router = in.Router
				},
			),
		)
	)

	require.NotNil(router)
	router.HandleFunc("/test", handler)
	app.RequireStart()
	app.RequireStop()
}

func TestUnmarshal(t *testing.T) {
	t.Run("Provide", func(t *testing.T) {
		t.Run("Full", testUnmarshalProvideFull)
		t.Run("Optional", testUnmarshalProvideOptional)
		t.Run("Required", testUnmarshalProvideRequired)
		t.Run("UnmarshalError", testUnmarshalProvideUnmarshalError)
		t.Run("ChainFactoryError", testUnmarshalProvideChainFactoryError)
	})

	t.Run("Annotated", func(t *testing.T) {
		t.Run("Full", testUnmarshalAnnotatedFull)
		t.Run("Named", testUnmarshalAnnotatedNamed)
	})
}
