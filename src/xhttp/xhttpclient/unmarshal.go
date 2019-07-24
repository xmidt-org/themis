package xhttpclient

import (
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type ClientIn struct {
	fx.In

	Viper *viper.Viper
}

// Unmarshal returns an uber/fx provider than in turn unmarshals client options
// and produces a client object.  If multiple client objects need to coexist
// in the same uber/fx App, use fx.Annotated with this function.
func Unmarshal(configKey string) func(ClientIn) (Interface, error) {
	return func(in ClientIn) (Interface, error) {
		var o Options
		if err := in.Viper.UnmarshalKey(configKey, &o); err != nil {
			return nil, err
		}

		return New(o)
	}
}
