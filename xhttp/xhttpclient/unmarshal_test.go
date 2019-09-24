package xhttpclient

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xmidt-org/themis/config"
	"github.com/xmidt-org/themis/xlog"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func testUnmarshalSuccess(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		handler = http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			assert.Equal("value", request.Header.Get("X-Header"))
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

	response, err := c.Do(request)
	require.NoError(err)
	require.NotNil(response)
	assert.Equal(299, response.StatusCode)
}

func testUnmarshalFailure(t *testing.T) {
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

func TestUnmarshal(t *testing.T) {
	t.Run("Success", testUnmarshalSuccess)
	t.Run("Failure", testUnmarshalFailure)
}
