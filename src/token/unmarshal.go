package token

import (
	"config"
	"key"
	"random"
	"xhttp/xhttpclient"

	"go.uber.org/fx"
)

type TokenIn struct {
	fx.In

	Noncer       random.Noncer
	Keys         key.Registry
	Unmarshaller config.Unmarshaller
	Client       xhttpclient.Interface `optional:"true"`
}

type TokenOut struct {
	fx.Out

	ClaimBuilder  ClaimBuilder
	Factory       Factory
	IssueHandler  IssueHandler
	ClaimsHandler ClaimsHandler
}

// Unmarshal returns an uber/fx style factory that produces the relevant components for
// a single token factory.
func Unmarshal(configKey string, b ...RequestBuilder) func(TokenIn) (TokenOut, error) {
	return func(in TokenIn) (TokenOut, error) {
		var o Options
		if err := in.Unmarshaller.UnmarshalKey(configKey, &o); err != nil {
			return TokenOut{}, err
		}

		cb, err := NewClaimBuilders(in.Noncer, in.Client, o)
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
			ClaimsHandler: NewClaimsHandler(
				NewClaimsEndpoint(cb),
				rb,
			),
		}, nil
	}
}
