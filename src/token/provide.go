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

	Claimer Claimer
	Factory Factory
	Handler Handler
}

func Provide(configKey string, b ...RequestBuilder) func(TokenIn) (TokenOut, error) {
	return func(in TokenIn) (TokenOut, error) {
		var o Options
		if err := in.Viper.UnmarshalKey(configKey, &o); err != nil {
			return TokenOut{}, err
		}

		c, err := NewClaimers(in.Noncer, o)
		if err != nil {
			return TokenOut{}, err
		}

		f, err := NewFactory(c, in.Keys, o)
		if err != nil {
			return TokenOut{}, err
		}

		b = append(b, NewTokenRequestBuilders(o)...)

		return TokenOut{
			Claimer: c,
			Factory: f,
			Handler: NewHandler(
				NewServerEndpoint(f),
				b...,
			),
		}, nil
	}
}
