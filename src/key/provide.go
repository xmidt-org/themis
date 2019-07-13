package key

import (
	"io"

	"go.uber.org/fx"
)

type KeyIn struct {
	fx.In

	Random io.Reader
}

type KeyOut struct {
	fx.Out

	Registry Registry
	Handler  Handler
}

func Provide(in KeyIn) KeyOut {
	registry := NewRegistry(in.Random)

	return KeyOut{
		Registry: registry,
		Handler: NewHandler(
			NewEndpoint(registry),
		),
	}
}
