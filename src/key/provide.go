package key

import (
	"io"

	"go.uber.org/fx"
)

// KeyIn is the set of dependencies for this package's components
type KeyIn struct {
	fx.In

	// Random is the optional source of randomness.  If not present in the container,
	// crypto/rand.Reader is used.
	Random io.Reader `optional:"true"`
}

// KeyOut is the set of components emitted by this package
type KeyOut struct {
	fx.Out

	// Registry is the fully configured token Registry
	Registry Registry

	// Handler is the http.Handler which can serve key requests to the Registry
	Handler Handler
}

// Provide is an uber/fx style provider for this package's components
func Provide(in KeyIn) KeyOut {
	registry := NewRegistry(in.Random)

	return KeyOut{
		Registry: registry,
		Handler: NewHandler(
			NewEndpoint(registry),
		),
	}
}
