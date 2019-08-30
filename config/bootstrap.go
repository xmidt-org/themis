package config

import (
	"os"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

// DiscardPrinter is an fx.Printer that simply ignores all logging.  Useful
// when errors need to short-circuit application startup, since in those situations
// the extra logging from the uber/fx App is just spam.
type DiscardPrinter struct{}

func (dp DiscardPrinter) Printf(string, ...interface{}) {
}

// Environment holds the boilerplate objects that describe or make up the application environment.
// An Initializer, if supplied, is invoked to setup the flagset (which includes parsing) and viper instances.
type Environment struct {
	Name         string
	Arguments    []string
	FlagSet      *pflag.FlagSet
	Viper        *viper.Viper
	Unmarshaller Unmarshaller
}

// Initializer is the strategy for initializing the application environment.  Implementations
// are responsible for configuring command line options, parsing the command line, and reading
// in viper configuration.
type Initializer func(Environment) error

// Bootstrap describes how to bootstrap an uber/fx application using the spf13/pflag and spf13/viper libraries.
// Since certain components need to be created prior to the uber/fx dependency injection flow, this type manages
// a simple workflow for application code to setup these components.  One example of such a component is logging,
// since most of the time the logging infrastructure should also be used for uber/fx's Logger option.
//
// All fields are optional.  The zero value of this type will simply bootstrap the standard command-line environment
// with all the defaults for the spf13/pflag and spf13/viper packages.
type Bootstrap struct {
	// Name is the application name, typically the executable name.  If unset, os.Args[0] is used.
	// The name is what is passed to NewFlagSet and is the same value passed to Initialize.
	Name string

	// Arguments are the command-line arguments to be parsed via the spf13/pflag package.  If unset, os.Args[1:] is used.
	Arguments []string

	// DisableDiscardOnError controls whether uber/fx logging output is discarded if an Initializer error occurs.
	DisableDiscardOnError bool

	// ErrorHandling is the pflag error handling strategy.  By default, this is ContinueOnError.
	ErrorHandling pflag.ErrorHandling

	// DecodeOptions are the optional Viper options for unmarshalling
	DecodeOptions []viper.DecoderConfigOption
}

// Provide performs initialization external to the uber/fx App flow, creating the various environmental
// components that need to exist prior to any providers running.  The pflag and viper instances are supplied as components, along
// with any components created by an Optioner strategy.
//
// The initializer parameter is the strategy used to setup the pflag instance, which includes parsing the command line,
// and read in any necessary viper configuration.  Any error returned by the Initializer causes application
// startup to be short-circuited with the error, which will be availabel via App.Err().  If the initializer is nil,
// the flagset and viper instances emitted into the uber/fx application are uninitialized.
func (b Bootstrap) Provide(i Initializer, optioners ...Optioner) fx.Option {
	name := b.Name
	if len(name) == 0 {
		name = os.Args[0]
	}

	arguments := b.Arguments
	if arguments == nil {
		arguments = os.Args[1:]
	}

	var (
		v = viper.New()
		e = Environment{
			Name:         name,
			Arguments:    arguments,
			FlagSet:      pflag.NewFlagSet(name, b.ErrorHandling),
			Viper:        v,
			Unmarshaller: ViperUnmarshaller{Viper: v, Options: b.DecodeOptions},
		}
	)

	if i != nil {
		if err := i(e); err != nil {
			if b.DisableDiscardOnError {
				return fx.Error(err)
			} else {
				return fx.Options(
					fx.Logger(DiscardPrinter{}),
					fx.Error(err),
				)
			}
		}
	}

	options := []fx.Option{
		fx.Provide(
			func() (*pflag.FlagSet, *viper.Viper, Unmarshaller) {
				return e.FlagSet, e.Viper, e.Unmarshaller
			},
		),
	}

	for _, f := range optioners {
		options = append(options, f(e))
	}

	return fx.Options(options...)
}
