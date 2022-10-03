package xhttpclient

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xmidt-org/candlelight"
	"github.com/xmidt-org/themis/config"
	"github.com/xmidt-org/themis/xlog"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func testUnmarshalProvideFull(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		handler = http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			assert.Equal("value", request.Header.Get("X-Header"))
			assert.Equal("from unmarshal chain", request.Header.Get("X-Unmarshal-Chain"))
			assert.Equal("from component chain", request.Header.Get("X-Component-Chain"))
			assert.Equal("from component chain factory", request.Header.Get("X-Component-Chain-Factory"))
			response.WriteHeader(299)
		})

		c   Interface
		app = fxtest.New(t,
			fx.Provide(
				config.ProvideViper(
					config.Json(`
						{
							"client": {
								"transport": {
									"idleConnTimeout": "1s",
									"tls": {
										"insecureSkipVerify": true
									}
								},
								"timeout": "10s",
								"header": {
									"x-header": ["value"]
								}
							}
						}
					`),
				),
				func() Chain {
					return NewChain(
						func(delegate http.RoundTripper) http.RoundTripper {
							return RoundTripperFunc(func(request *http.Request) (*http.Response, error) {
								request.Header.Set("X-Component-Chain", "from component chain")
								return delegate.RoundTrip(request)
							})
						},
					)
				},
				func() ChainFactory {
					return ChainFactoryFunc(func(name string, o Options) (Chain, error) {
						assert.Equal("client", name)
						return NewChain(
							func(delegate http.RoundTripper) http.RoundTripper {
								return RoundTripperFunc(func(request *http.Request) (*http.Response, error) {
									request.Header.Set("X-Component-Chain-Factory", "from component chain factory")
									return delegate.RoundTrip(request)
								})
							},
						), nil
					})
				},
				Unmarshal{
					Key: "client",
					Chain: NewChain(
						func(delegate http.RoundTripper) http.RoundTripper {
							return RoundTripperFunc(func(request *http.Request) (*http.Response, error) {
								request.Header.Set("X-Unmarshal-Chain", "from unmarshal chain")
								return delegate.RoundTrip(request)
							})
						},
					),
				}.Provide,
				func() (candlelight.Tracing, error) {
					return candlelight.New(candlelight.Config{Provider: "stdout"})
				},
			),
			fx.Populate(&c),
		)
	)

	require.NoError(app.Err())
	require.NotNil(c)

	s := httptest.NewServer(handler)
	defer s.Close()

	request, err := http.NewRequest("GET", s.URL, nil)
	require.NoError(err)

	response, err := c.Do(request) //nolint:bodyclose
	require.NoError(err)
	require.NotNil(response)
	assert.Equal(299, response.StatusCode)
}

func testUnmarshalProvideWithRoundTripper(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		handler = http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			assert.Equal("value", request.Header.Get("X-Header"))
			assert.Equal("from roundtripper component", request.Header.Get("X-RoundTripper-Component"))
			response.WriteHeader(299)
		})

		c   Interface
		app = fxtest.New(t,
			fx.Provide(
				config.ProvideViper(
					config.Json(`
						{
							"client": {
								"transport": {
									"idleConnTimeout": "1s",
									"tls": {
										"insecureSkipVerify": true
									}
								},
								"timeout": "10s",
								"header": {
									"x-header": ["value"]
								}
							}
						}
					`),
				),
				func() http.RoundTripper {
					return NewChain(
						func(delegate http.RoundTripper) http.RoundTripper {
							return RoundTripperFunc(func(request *http.Request) (*http.Response, error) {
								request.Header.Set("X-RoundTripper-Component", "from roundtripper component")
								return delegate.RoundTrip(request)
							})
						},
					).Then(nil)
				},
				Unmarshal{Key: "client"}.Provide,
			),
			fx.Populate(&c),
		)
	)

	require.NoError(app.Err())
	require.NotNil(c)

	s := httptest.NewServer(handler)
	defer s.Close()

	request, err := http.NewRequest("GET", s.URL, nil)
	require.NoError(err)

	response, err := c.Do(request) //nolint:bodyclose
	require.NoError(err)
	require.NotNil(response)
	assert.Equal(299, response.StatusCode)
}

func testUnmarshalProvideUnmarshalError(t *testing.T) {
	var (
		assert = assert.New(t)

		c   Interface
		app = fx.New(
			fx.Logger(xlog.DiscardPrinter{}),
			fx.Provide(
				config.ProvideViper(
					config.Json(`
							{
								"client": {
									"timeout": "this is not a valid duration"
								}
							}
					`),
				),
				Unmarshal{Key: "client"}.Provide,
			),
			fx.Populate(&c),
		)
	)

	assert.Error(app.Err())
	assert.Nil(c)
}

func testUnmarshalProvideChainFactoryError(t *testing.T) {
	var (
		assert      = assert.New(t)
		expectedErr = errors.New("expected chain factory error")

		app = fx.New(
			fx.Logger(xlog.DiscardPrinter{}),
			fx.Provide(
				xlog.Provide(log.NewNopLogger()),
				config.ProvideViper(
					config.Json(`
						{
							"client": {
								"transport": {
									"idleConnTimeout": "1s"
								}
							}
						}
					`),
				),
				func() ChainFactory {
					return ChainFactoryFunc(func(name string, o Options) (Chain, error) {
						return Chain{}, expectedErr
					})
				},
				Unmarshal{Key: "client"}.Provide,
			),
			fx.Invoke(
				func(Interface) {
					assert.Fail("This invoke function should not have been called")
				},
			),
		)
	)

	assert.Error(app.Err())
}

type testUnmarshalAnnotatedFullIn struct {
	fx.In

	Client Interface `name:"client" optional:"true"`
}

func testUnmarshalAnnotatedFull(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		handler = http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			assert.Equal("value", request.Header.Get("X-Header"))
			assert.Equal("from unmarshal chain", request.Header.Get("X-Unmarshal-Chain"))
			assert.Equal("from component chain", request.Header.Get("X-Component-Chain"))
			assert.Equal("from component chain factory", request.Header.Get("X-Component-Chain-Factory"))
			response.WriteHeader(299)
		})

		c   Interface
		app = fxtest.New(t,
			fx.Provide(
				config.ProvideViper(
					config.Json(`
						{
							"client": {
								"transport": {
									"idleConnTimeout": "1s",
									"tls": {
										"insecureSkipVerify": true
									}
								},
								"timeout": "10s",
								"header": {
									"x-header": ["value"]
								}
							}
						}
					`),
				),
				func() Chain {
					return NewChain(
						func(delegate http.RoundTripper) http.RoundTripper {
							return RoundTripperFunc(func(request *http.Request) (*http.Response, error) {
								request.Header.Set("X-Component-Chain", "from component chain")
								return delegate.RoundTrip(request)
							})
						},
					)
				},
				func() ChainFactory {
					return ChainFactoryFunc(func(name string, o Options) (Chain, error) {
						assert.Equal("client", name)
						return NewChain(
							func(delegate http.RoundTripper) http.RoundTripper {
								return RoundTripperFunc(func(request *http.Request) (*http.Response, error) {
									request.Header.Set("X-Component-Chain-Factory", "from component chain factory")
									return delegate.RoundTrip(request)
								})
							},
						), nil
					})
				},
				Unmarshal{
					Key: "client",
					Chain: NewChain(
						func(delegate http.RoundTripper) http.RoundTripper {
							return RoundTripperFunc(func(request *http.Request) (*http.Response, error) {
								request.Header.Set("X-Unmarshal-Chain", "from unmarshal chain")
								return delegate.RoundTrip(request)
							})
						},
					),
				}.Annotated(),
			),
			fx.Invoke(
				func(in testUnmarshalAnnotatedFullIn) {
					assert.NotNil(in.Client)
					c = in.Client
				},
			),
		)
	)

	require.NoError(app.Err())
	require.NotNil(c)

	s := httptest.NewServer(handler)
	defer s.Close()

	request, err := http.NewRequest("GET", s.URL, nil)
	require.NoError(err)

	response, err := c.Do(request) //nolint: bodyclose
	require.NoError(err)
	require.NotNil(response)
	assert.Equal(299, response.StatusCode)
}

type testUnmarshalAnnotatedNamedIn struct {
	fx.In

	Client Interface `name:"componentName" optional:"true"`
}

func testUnmarshalAnnotatedNamed(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		c   Interface
		app = fxtest.New(t,
			fx.Provide(
				xlog.Provide(log.NewNopLogger()),
				config.ProvideViper(
					config.Json(`
						{
							"client": {
								"transport": {
									"idleConnTimeout": "1s",
									"tls": {
										"insecureSkipVerify": true
									}
								},
								"timeout": "10s"
							}
						}
					`),
				),
				Unmarshal{
					Key:  "server",
					Name: "componentName",
				}.Annotated(),
			),
			fx.Invoke(
				func(in testUnmarshalAnnotatedNamedIn) {
					assert.NotNil(in.Client)
					c = in.Client
				},
			),
		)
	)

	require.NotNil(c)
	app.RequireStart()
	app.RequireStop()
}

func TestUnmarshal(t *testing.T) {
	t.Run("Provide", func(t *testing.T) {
		t.Run("Full", testUnmarshalProvideFull)
		t.Run("WithRoundTripper", testUnmarshalProvideWithRoundTripper)
		t.Run("UnmarshalError", testUnmarshalProvideUnmarshalError)
		t.Run("ChainFactoryError", testUnmarshalProvideChainFactoryError)
	})

	t.Run("Annotated", func(t *testing.T) {
		t.Run("Full", testUnmarshalAnnotatedFull)
		t.Run("Named", testUnmarshalAnnotatedNamed)
	})
}
