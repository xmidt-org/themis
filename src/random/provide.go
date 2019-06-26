package random

import (
	"crypto/rand"
	"encoding/base64"
	"io"

	"go.uber.org/fx"
)

// Out describes the components emitted by this package
type Out struct {
	fx.Out

	Random io.Reader
	Noncer Noncer
}

// Provide is a DI provider that creates this package's components
func Provide() Out {
	return Out{
		Random: rand.Reader,
		Noncer: NewBase64Noncer(rand.Reader, 16, base64.RawURLEncoding),
	}
}
