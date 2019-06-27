package key

import (
	"io"

	"go.uber.org/fx"
)

type In struct {
	fx.In

	Random io.Reader
}

type Out struct {
	fx.Out

	Registry Registry
	Handler  Handler
}

func Provide(in In) Out {
	registry := NewRegistry(in.Random)

	return Out{
		Registry: registry,
		Handler: NewHandler(
			NewEndpoint(registry),
		),
	}
}
