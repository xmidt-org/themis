package xhttpclient

import (
	"net/http"

	"github.com/xmidt-org/themis/config"

	"go.uber.org/fx"
)

// ClientUnmarshalIn defines the set of dependencies for an HTTP client
type ClientUnmarshalIn struct {
	fx.In

	// Unmarshaller is the required configuration unmarshalling component
	Unmarshaller config.Unmarshaller

	// Chain is an optional component.  If present in the application, this chain
	// will be used to decorate the roundtripper.  Any constructors set via Unmarshal
	// will be appended to this chain.  This field and ChainFactory can be used together, and both
	// will be used to produce a merged Chain if both fields are set.
	//
	// Using this component allows for global decorators that apply to all clients.
	Chain Chain `optional:"true"`

	// ChainFactory is an optional component used to create a Chain specific to a given HTTP client.
	// Both this field and Chain may be set, in which case both are used to create a single Chain.
	ChainFactory ChainFactory `optional:"true"`

	// RoundTripper is an optional http.RoundTripper component.  If present, this field will be used
	// for clients unmarshalled by this instance.  Configuration will be ignored in favor of this component.
	RoundTripper http.RoundTripper `optional:"true"`
}

// Unmarshal encompasses all the non-component information for unmarshalling and instantiating
// an HTTP client.
type Unmarshal struct {
	// Key is the required configuration key unmarshalled via Viper.  This key is unmarshalled as an Options.
	Key string

	// Name is the optional name of this client within the application.  If unset, the Key is used.  This
	// field is used when providing a named client component via Annotated.
	Name string

	// Chain is an optional decoration for the client's RoundTripper.  This field is used for decoration
	// initialized outside the uber/fx application.
	Chain Chain
}

func (u Unmarshal) name() string {
	if len(u.Name) > 0 {
		return u.Name
	}

	return u.Key
}

// Provide emits an HTTP client as an unnamed component.  If only (1) client is needed for the entire
// application, this method is best.
func (u Unmarshal) Provide(in ClientUnmarshalIn) (Interface, error) {
	var o Options
	if err := in.Unmarshaller.UnmarshalKey(u.Key, &o); err != nil {
		return nil, err
	}

	var rt http.RoundTripper
	if in.RoundTripper != nil {
		rt = in.RoundTripper
	} else {
		rt = NewRoundTripper(o.Transport)
	}

	chain := in.Chain.Extend(u.Chain)
	if in.ChainFactory != nil {
		more, err := in.ChainFactory.NewClientChain(u.name(), o)
		if err != nil {
			return nil, err
		}

		chain = chain.Extend(more)
	}

	return NewCustom(o, chain.Then(rt)), nil
}

// Annotated returns an uber/fx Annotated instance that emits an HTTP client with a specific name.
// The Name field is used if set, otherwise Key is used.  If multiple HTTP clients are needed for the
// enclosing application, use this function to emit them as named components.
func (u Unmarshal) Annotated() fx.Annotated {
	return fx.Annotated{
		Name:   u.name(),
		Target: u.Provide,
	}
}
