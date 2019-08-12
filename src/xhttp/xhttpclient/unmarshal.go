package xhttpclient

import (
	"config"

	"go.uber.org/fx"
)

// ClientUnmarshalIn defines the set of dependencies for an HTTP client
type ClientUnmarshalIn struct {
	fx.In

	Unmarshaller config.Unmarshaller

	// Chain is an optional component.  If present in the application, this chain
	// will be used to decorate the roundtripper.  Any constructors passed to Unmarshal
	// will be appended to this chain.
	//
	// Using this component allows for global decorators that apply to all clients.
	Chain Chain `optional:"true"`
}

// Unmarshal returns an uber/fx provider than in turn unmarshals client options
// and produces a client object.  If multiple client objects need to coexist
// in the same uber/fx App, use fx.Annotated with this function.
func Unmarshal(configKey string, c ...Constructor) func(ClientUnmarshalIn) (Interface, error) {
	return func(in ClientUnmarshalIn) (Interface, error) {
		var o Options
		if err := in.Unmarshaller.UnmarshalKey(configKey, &o); err != nil {
			return nil, err
		}

		return New(o, in.Chain.Append(c...).Then), nil
	}
}
