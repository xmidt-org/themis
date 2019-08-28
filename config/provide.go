package config

import (
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

// ViperOut lists the components emitted for a Viper instance
type ViperOut struct {
	fx.Out

	Viper        *viper.Viper
	Unmarshaller Unmarshaller
}

// ProvideViper emits the various uber/fx components related to Viper.  This provider can
// be used standalone.  It is used by Bootstrap.Provide() as well.
func ProvideViper(v *viper.Viper, options ...viper.DecoderConfigOption) func() ViperOut {
	return func() ViperOut {
		return ViperOut{
			Viper:        v,
			Unmarshaller: ViperUnmarshaller{Viper: v, Options: options},
		}
	}
}
