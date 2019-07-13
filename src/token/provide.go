package token

import (
	"key"
	"random"

	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type TokenIn struct {
	fx.In

	Noncer random.Noncer
	Keys   key.Registry
	Viper  *viper.Viper
}

type TokenOut struct {
	fx.Out

	Factory Factory
	Handler Handler
}

func Provide(configKey string, b ...TokenRequestBuilder) func(TokenIn) (TokenOut, error) {
	return func(in TokenIn) (TokenOut, error) {
		var d Descriptor
		if err := in.Viper.UnmarshalKey(configKey, &d); err != nil {
			return TokenOut{}, err
		}

		f, err := NewFactory(in.Noncer, in.Keys, d)
		if err != nil {
			return TokenOut{}, err
		}

		b = append(b, NewTokenRequestBuilders(d)...)

		return TokenOut{
			Factory: f,
			Handler: NewHandler(
				NewServerEndpoint(f),
				b...,
			),
		}, nil
	}
}
