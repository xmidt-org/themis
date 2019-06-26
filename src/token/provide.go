package token

import (
	"key"
	"random"

	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type In struct {
	fx.In

	Noncer random.Noncer
	Keys   key.Registry
	Viper  *viper.Viper
}

type Out struct {
	fx.Out

	Factory Factory
}

func Provide(key string) func(In) (Out, error) {
	return func(in In) (Out, error) {
		var d Descriptor
		if err := in.Viper.UnmarshalKey(key, &d); err != nil {
			return Out{}, err
		}

		f, err := NewFactory(in.Noncer, in.Keys, d)
		if err != nil {
			return Out{}, err
		}

		return Out{
			Factory: f,
		}, nil
	}
}
