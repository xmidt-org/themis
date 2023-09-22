// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0 
package config

import (
	"io"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

// ViperIn describes the dependencies for creating an spf13 Viper instance
type ViperIn struct {
	fx.In

	// Name is the application (or, executable) name.  This component will be supplied if using
	// CommandLine.Provide.  If unset, DefaultApplicationName is used.  The instance of this struct
	// passed to builders will always have this field set.
	//
	// This name is frequently used to determine names of or paths to configuration files.  It is
	// passed to each builder function in Viper.Provide.
	Name ApplicationName `optional:"true"`

	// FlagSet is the optional spf13 FlagSet component.  It is passed to each builder function.
	// If supplied, this component can be used to tailor the viper instance using command-line arguments.
	FlagSet *pflag.FlagSet `optional:"true"`

	// DecoderOptions is an optional component that is a slice of spf13/viper decoder options for unmarshalling.
	// If supplied, this slice is used to create the Unmarshaller component.
	//
	// Note that spf13/viper provides a default set of options.  See https://godoc.org/github.com/spf13/viper#DecoderConfigOption
	DecoderOptions []viper.DecoderConfigOption `optional:"true"`
}

// ViperOut lists the components emitted for a Viper instance
type ViperOut struct {
	fx.Out

	Viper        *viper.Viper
	Unmarshaller Unmarshaller
}

// ViperBuilder is a builder strategy for tailoring a viper instance.  The ViperIn set of dependencies will
// always have the Name field set, either as a component or by defaulting to DefaultApplicationName.
type ViperBuilder func(ViperIn, *viper.Viper) error

// ProvideViper produces components for the viper environment.  Each builder is invoked in sequence, and any error
// will short-circuit construction.  This provider function does not itself read any configuration or otherwise
// modify the viper instance it creates.  At least one builder function must read configuration.
func ProvideViper(builders ...ViperBuilder) func(ViperIn) (ViperOut, error) {
	return func(in ViperIn) (ViperOut, error) {
		if len(in.Name) == 0 {
			in.Name = DefaultApplicationName()
		}

		viper := viper.New()
		for _, f := range builders {
			if err := f(in, viper); err != nil {
				return ViperOut{}, err
			}
		}

		return ViperOut{
			Viper:        viper,
			Unmarshaller: ViperUnmarshaller{Viper: viper, Options: in.DecoderOptions},
		}, nil
	}
}

// ReadConfig returns a ViperBuilder that reads in a configuration file from an arbitrary io.Reader.
func ReadConfig(format string, data io.Reader) ViperBuilder {
	return func(_ ViperIn, v *viper.Viper) error {
		v.SetConfigType(format)
		return v.ReadConfig(data)
	}
}

// Json uses ReadConfig to read configuration from an in-memory JSON string.
func Json(data string) ViperBuilder {
	return ReadConfig("json", strings.NewReader(data))
}

// Yaml uses ReadConfig to read configuration from an in-memory YAML string.
func Yaml(data string) ViperBuilder {
	return ReadConfig("yaml", strings.NewReader(data))
}
