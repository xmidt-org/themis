package xloghttp

import (
	"github.com/go-kit/kit/log"
	"go.uber.org/fx"
)

type ProvideIn struct {
	fx.In

	Base     log.Logger
	Builders []ParameterBuilder
}

func Provide(in ProvideIn) Logging {
	return Logging{
		Base:     in.Base,
		Builders: in.Builders,
	}
}
