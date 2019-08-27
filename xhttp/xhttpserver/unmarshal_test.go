package xhttpserver

import (
	"net/http"
	"testing"

	"github.com/xmidt-org/themis/config"
	"github.com/xmidt-org/themis/config/configtest"
	"github.com/xmidt-org/themis/xlog"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func testUnmarshalFull(t *testing.T) {
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
				xlog.Provide(xlog.Discard()),
				config.ProvideViper(
					configtest.LoadJson(t,
						`
						{
							"server": {
								"address": "127.0.0.1:0",
								"disableHTTPKeepAlives": true
							}
						}
					`,
					),
				),
				func() alice.Chain {
					return alice.New(
						ResponseHeaders{
							Header: http.Header{"X-Global": []string{"global value"}},
						}.Then,
					)
				},
				func() ChainFactory {
					return ChainFactoryFunc(func(o Options) (alice.Chain, error) {
						return alice.New(
							ResponseHeaders{
								Header: http.Header{"X-Factory": []string{o.Name}},
							}.Then,
						), nil
					})
				},
				Unmarshal(
					"server",
					ResponseHeaders{
						Header: http.Header{"X-Adhoc": []string{"adhoc value"}},
					}.Then,
				),
			),
			fx.Populate(&router),
		)
	)

	require.NotNil(router)
	router.HandleFunc("/test", handler)
	app.RequireStart()
	app.RequireStop()
}

func TestUnmarshal(t *testing.T) {
	// TODO: add event code that can determine the host:port of the server
	t.Run("Full", testUnmarshalFull)
}
