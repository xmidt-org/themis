package key

import (
	"io"
	"xhttp/xhttpserver"

	"github.com/gorilla/mux"
	"go.uber.org/fx"
)

type ProvideIn struct {
	fx.In

	Random io.Reader
}

type ProvideOut struct {
	fx.Out

	Registry Registry
	Handler  Handler
	Router   *mux.Router `name:"keyRouter"`
}

func Provide(serverConfigKey string) func(ProvideIn, xhttpserver.ProvideIn) (ProvideOut, error) {
	return func(keyIn ProvideIn, serverIn xhttpserver.ProvideIn) (ProvideOut, error) {
		keyRouter, err := xhttpserver.Provide(serverConfigKey)(serverIn)
		if err != nil {
			return ProvideOut{}, err
		}

		registry := NewRegistry(keyIn.Random)

		return ProvideOut{
			Registry: registry,
			Handler: NewHandler(
				NewEndpoint(registry),
			),
			Router: keyRouter,
		}, nil
	}
}

type InvokeIn struct {
	fx.In

	Handler Handler
	Router  *mux.Router `name:"keyRouter"`
}

func RunServer(path string) func(InvokeIn) {
	return func(in InvokeIn) {
		in.Router.Handle(path, in.Handler)
	}
}
