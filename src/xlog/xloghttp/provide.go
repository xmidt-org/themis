package xloghttp

import (
	"github.com/go-kit/kit/log"
	"go.uber.org/fx"
)

type ProvideIn struct {
	fx.In

	Logger log.Logger
}

type ProvideOut struct {
	fx.Out

	Constructor Constructor
}

func Provide(b ...ParameterBuilder) func(ProvideIn) ProvideOut {
	return func(in ProvideIn) ProvideOut {
		return ProvideOut{
			Constructor: NewConstructor(in.Logger, b...),
		}
	}
}
