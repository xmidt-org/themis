package token

import (
	"key"
	"random"
	"xhttp/xhttpserver"

	"github.com/gorilla/mux"
	"go.uber.org/fx"
)

type ProvideIn struct {
	fx.In

	Noncer random.Noncer
	Keys   key.Registry
}

type ProvideOut struct {
	fx.Out

	Factory Factory
	Handler Handler
	Router  *mux.Router `name:"tokenRouter"`
}

func Provide(serverConfigKey, tokenConfigKey string, b ...RequestBuilder) func(ProvideIn, xhttpserver.ProvideIn) (ProvideOut, error) {
	return func(in ProvideIn, serverIn xhttpserver.ProvideIn) (ProvideOut, error) {
		router, err := xhttpserver.Unmarshal(serverConfigKey, serverIn)
		if err != nil {
			return ProvideOut{}, err
		}

		var d Descriptor
		if err := serverIn.Viper.UnmarshalKey(tokenConfigKey, &d); err != nil {
			return ProvideOut{}, err
		}

		f, err := NewFactory(in.Noncer, in.Keys, d)
		if err != nil {
			return ProvideOut{}, err
		}

		b = append(b, NewBuilders(d)...)

		return ProvideOut{
			Factory: f,
			Handler: NewHandler(
				NewEndpoint(f),
				b...,
			),
			Router: router,
		}, nil
	}
}

type InvokeIn struct {
	fx.In

	Handler Handler
	Router  *mux.Router `name:"tokenRouter"`
}

func RunServer(path string) func(in InvokeIn) {
	return func(in InvokeIn) {
		in.Router.Handle(path, in.Handler)
	}
}
