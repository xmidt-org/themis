package xconfig

import (
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type ViperOut struct {
	fx.Out

	Viper           *viper.Viper
	Unmarshaller    Unmarshaller
	KeyUnmarshaller KeyUnmarshaller
}

// ProvideViper emits the various uber/fx components related to Viper.
func ProvideViper(v *viper.Viper, options ...viper.DecoderConfigOption) func() ViperOut {
	return func() ViperOut {
		return ViperOut{
			Viper:           v,
			Unmarshaller:    ViperUnmarshaller{Viper: v, Options: options},
			KeyUnmarshaller: ViperUnmarshaller{Viper: v, Options: options},
		}
	}
}
