// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0 
package config

import (
	"os"

	"github.com/spf13/pflag"
	"go.uber.org/fx"
)

// ApplicationName is the type of component that identifies the executable or application name.
type ApplicationName string

// DefaultApplicationName simply returns os.Arg[0]
func DefaultApplicationName() ApplicationName {
	return ApplicationName(os.Args[0])
}

// CommandLineOut defines the components provided by CommandLine.Provide
type CommandLineOut struct {
	fx.Out

	// Name is the application (or, executable) name determined by CommandLine.Provide.
	// It defaults to os.Args[0].  This component will always be supplied regardless of
	// the CommandLine settings.
	Name ApplicationName

	// FlagSet is the command line flags and configuration.  This will be parsed by default.
	FlagSet *pflag.FlagSet
}

// CommandLine describes how to provide a *pflag.FlagSet to an uber/fx container.  The zero value
// for this type is valid and will parse the executable's command line defined by os.Args.  Examples include:
//
//	CommandLine{}.Provide(parseCommandLine)
//	CommandLine{Name: "custom"}.Provide(builder1, builder2)
//	CommandLine{DisableParse: true}.Provide(parseIt)
type CommandLine struct {
	// Name is the executable or application name.  If unset, os.Args[0] is used.
	Name string

	// Arguments are the command-line arguments to parse.  If nil, os.Args[1:] is used.  Note that
	// if this slice is explicitly set to an empty slice, then an empty slice of arguments will be parsed.
	Arguments []string

	// ErrorHandling is the pflag error handling mode.  Defaults to ContinueOnError.
	ErrorHandling pflag.ErrorHandling

	// DisableParse controls whether Provide will parse the command line after building the flagset.
	// Setting this to true allows upstream providers or invoke functions to handle the parsing.
	DisableParse bool
}

// FlagSetBuilder is a builder strategy for tailoring a flagset.  Most commonly, this involves
// setting up any command line flags.
type FlagSetBuilder func(*pflag.FlagSet) error

// Provide is an uber/fx provide function that creates a *pflag.FlagSet.  Zero or more builders may be
// passed to configure flagset, which includes adding command-line arguments.  If none of the builders
// parse the command line, and if DisableParse is false, this function will parse the arguments
// prior to returning the flagset.
//
// This provider will short-circuit application startup using fx.Error if any command-line parsing error occurs.
func (cl CommandLine) Provide(builders ...FlagSetBuilder) fx.Option {
	name := os.Args[0]
	if len(cl.Name) > 0 {
		name = cl.Name
	}

	fs := pflag.NewFlagSet(name, cl.ErrorHandling)
	var builderErrs []error
	for _, f := range builders {
		if err := f(fs); err != nil {
			builderErrs = append(builderErrs, err)
		}
	}

	if len(builderErrs) > 0 {
		return fx.Error(builderErrs...)
	}

	if !cl.DisableParse && !fs.Parsed() {
		arguments := cl.Arguments
		if arguments == nil {
			arguments = os.Args[1:]
		}

		if err := fs.Parse(arguments); err != nil {
			return fx.Error(err)
		}
	}

	return fx.Provide(
		func() CommandLineOut {
			return CommandLineOut{
				Name:    ApplicationName(name),
				FlagSet: fs,
			}
		},
	)
}
