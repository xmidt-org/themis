package issuer

import (
	"token"
	"xhttp/xhttpserver"

	"github.com/gorilla/mux"
	"go.uber.org/fx"
)

type ProvideIn struct {
	fx.In

	Factory token.Factory
}

type ProvideOut struct {
	fx.Out

	Issuer  Issuer
	Handler Handler
	Router  *mux.Router `name:"issuerRouter"`
}

func Provide(serverConfigKey, issuerConfigKey string) func(ProvideIn, xhttpserver.ProvideIn) (ProvideOut, error) {
	return func(issuerIn ProvideIn, serverIn xhttpserver.ProvideIn) (ProvideOut, error) {
		router, err := xhttpserver.Provide(serverConfigKey)(serverIn)
		if err != nil {
			return ProvideOut{}, err
		}

		var o Options
		if err := serverIn.Viper.UnmarshalKey(issuerConfigKey, &o); err != nil {
			return ProvideOut{}, err
		}

		i := New(issuerIn.Factory, o)

		return ProvideOut{
			Issuer: i,
			Handler: NewHandler(
				NewEndpoint(i),
			),
			Router: router,
		}, nil
	}
}

type InvokeIn struct {
	fx.In

	Handler Handler
	Router  *mux.Router `name:"issuerRouter"`
}

func RunServer(path string) func(in InvokeIn) {
	return func(in InvokeIn) {
		in.Router.Handle(path, in.Handler)
	}
}
