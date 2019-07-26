package xhttpclient

import (
	"xconfig"

	"go.uber.org/fx"
)

type UnmarshalIn struct {
	fx.In

	Unmarshaller xconfig.KeyUnmarshaller
}

// Unmarshal returns an uber/fx provider than in turn unmarshals client options
// and produces a client object.  If multiple client objects need to coexist
// in the same uber/fx App, use fx.Annotated with this function.
func Unmarshal(configKey string, c ...Constructor) func(UnmarshalIn) (Interface, error) {
	return func(in UnmarshalIn) (Interface, error) {
		var o Options
		if err := in.Unmarshaller.UnmarshalKey(configKey, &o); err != nil {
			return nil, err
		}

		return New(o, c...)
	}
}
