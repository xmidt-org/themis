package bootstrap

import (
	"errors"
	"os"
	"xlog"

	"github.com/go-kit/kit/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

var (
	ErrNoInitialize = errors.New("An Initialize closure is required")
)

// Environment supplies the externally created components and configurable options for bootstrap an fx application
// using spf13's flagset and viper libraries together with go-kit logging.
type Environment struct {
	// Name is the application name, typically the executable name.  If unset, os.Args[0] is used.
	Name string

	// Arguments are the command-line arguments to be parsed via the spf13/pflag package.  If unset, os.Args[1:] is used.
	Arguments []string

	// ErrorHandling is the pflag error handling strategy.  By default, this is ContinueOnError.
	ErrorHandling pflag.ErrorHandling

	// LogKey is the viper configuration key where logging configuration is supplied.
	// There is no default.  If unset, xlog.Default() is used as the logger.
	LogKey string

	// Initialize is the required closure used to initialize the environment.  This function should configure
	// the flagset and viper, parse the command line, and read in configuration as appropriate.  It represents
	// the application layer's specific code to bootstrap the environment.
	Initialize func(name string, arguments []string, fs *pflag.FlagSet, v *viper.Viper) error
}

// Options prepends the appropriate bootstrapping to a variadic set of uber/fx options.  The result
// is returned, and can be used via fx.New.
//
// This function does the following:
//   - Creates pflag.FlagSet and viper.Viper instances
//   - Invokes the Initialize closure to allow application code to setup the environment
//   - Unmarshals logging configuration from the viper instance and creates a go-kit logger
//   - Sets the logger as the uber/fx printer
//   - Provides the flagset, viper, and logger instances as application components
//
// Any errors that occur during bootstrapping are emitted as fx.Invoke functions and will be available via App.Err().
// For example, if the help options was requested on the command line, pflag.ErrHelp will be returned to the application.
func (e Environment) Options(opts ...fx.Option) []fx.Option {
	if e.Initialize == nil {
		return []fx.Option{
			fx.Logger(xlog.Printer{Logger: xlog.Discard()}),
			fx.Invoke(func() error { return ErrNoInitialize }),
		}
	}

	name := e.Name
	if len(name) == 0 {
		name = os.Args[0]
	}

	arguments := e.Arguments
	if arguments == nil {
		arguments = os.Args[1:]
	}

	var (
		fs = pflag.NewFlagSet(name, e.ErrorHandling)
		v  = viper.New()
	)

	if err := e.Initialize(name, arguments, fs, v); err != nil {
		return []fx.Option{
			fx.Logger(xlog.Printer{Logger: xlog.Discard()}),
			fx.Invoke(func() error { return err }),
		}
	}

	logger := xlog.Default()
	if len(e.LogKey) > 0 {
		var err error
		logger, err = xlog.Unmarshal(e.LogKey, v)
		if err != nil {
			return []fx.Option{
				fx.Logger(xlog.Printer{Logger: xlog.Discard()}),
				fx.Invoke(func() error { return err }),
			}
		}
	}

	return append(
		[]fx.Option{
			fx.Logger(xlog.Printer{Logger: logger}),
			fx.Provide(
				func() log.Logger { return logger },
				func() *pflag.FlagSet { return fs },
				func() *viper.Viper { return v },
			),
		},
		opts...,
	)
}
