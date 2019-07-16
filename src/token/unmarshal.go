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

	ClaimBuilder ClaimBuilder
	Factory      Factory
	IssueHandler IssueHandler
}

// Unmarshal returns an uber/fx style factory that produces the relevant components for
// a single token factory.
func Unmarshal(configKey string, b ...RequestBuilder) func(TokenIn) (TokenOut, error) {
	return func(in TokenIn) (TokenOut, error) {
		var o Options
		if err := in.Viper.UnmarshalKey(configKey, &o); err != nil {
			return TokenOut{}, err
		}

		cb, err := NewClaimBuilders(in.Noncer, o)
		if err != nil {
			return TokenOut{}, err
		}

		f, err := NewFactory(cb, in.Keys, o)
		if err != nil {
			return TokenOut{}, err
		}

		rb, err := NewRequestBuilders(o)
		if err != nil {
			return TokenOut{}, err
		}

		rb = append(rb, b...)
		return TokenOut{
			ClaimBuilder: cb,
			Factory:      f,
			IssueHandler: NewIssueHandler(
				NewIssueEndpoint(f),
				rb,
			),
		}, nil
	}
}
