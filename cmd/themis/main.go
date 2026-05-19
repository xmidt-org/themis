// SPDX-FileCopyrightText: 2019 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"syscall"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/xmidt-org/themis"
	"github.com/xmidt-org/themis/config"
	"go.uber.org/fx"
)

func app(done chan struct{}, startCtx, stopCtx context.Context, opt ...fx.Option) (sig syscall.Signal) {
	app, err := themis.New(
		fx.Options(
			append(opt,
				config.CommandLine{Name: themis.ApplicationName}.Provide(setupFlagSet),
				fx.Provide(
					fx.Annotate(func() config.ViperBuilder { return setupViper }, fx.ResultTags(`group:"viperBuilders"`)),
				),
			)...,
		),
	)
	if errors.Is(err, pflag.ErrHelp) {
		return 0
	} else if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return syscall.SIGINT
	}

	if err := app.Start(startCtx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return syscall.SIGTRAP
	}

	fxWait := app.Wait()
	var fxSig fx.ShutdownSignal
	select {
	case <-done:
		if err := stopCtx.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "using a new background context in order to attempt a graceful shutdown, previous stop context error'ed out: %s", err)
			stopCtx = context.Background()
		}

		if err := app.Stop(stopCtx); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return syscall.SIGHUP
		}

		fxSig = <-fxWait
	case fxSig = <-fxWait:
	}

	return syscall.Signal(fxSig.ExitCode)
}

func main() {
	os.Exit(int(app(nil, context.Background(), context.Background())))
}

func setupFlagSet(fs *pflag.FlagSet) error {
	fs.StringP("file", "f", "", "the configuration file to use.  Overrides the search path.")
	fs.Bool("dev", false, "development mode")
	fs.String("iss", "", "the name of the issuer to put into claims.  Overrides configuration.")
	fs.BoolP("debug", "d", false, "enables debug logging.  Overrides configuration.")
	fs.BoolP("version", "v", false, "print version and exit")

	return nil
}

func setupViper(in config.ViperIn, v *viper.Viper) error {
	if printVersion, _ := in.FlagSet.GetBool("version"); printVersion {
		return printVersionInfo()
	}

	var errs []error
	if dev, _ := in.FlagSet.GetBool("dev"); dev {
		v.SetConfigType("yaml")
		errs = append(errs, v.ReadConfig(strings.NewReader(devMode)))
	} else if file, _ := in.FlagSet.GetString("file"); len(file) > 0 {
		v.SetConfigFile(file)
		errs = append(errs, v.ReadInConfig())
	} else {
		v.SetConfigName(string(in.Name))
		v.AddConfigPath(fmt.Sprintf("/etc/%s", in.Name))
		v.AddConfigPath(".")
		v.AddConfigPath(fmt.Sprintf("$HOME/.%s", in.Name))
		errs = append(errs, v.ReadInConfig())
	}

	if iss, _ := in.FlagSet.GetString("iss"); len(iss) > 0 {
		v.Set("issuer.claims.iss", iss)
	}

	if debug, _ := in.FlagSet.GetBool("debug"); debug {
		v.Set("log.level", "DEBUG")
	}

	return errors.Join(errs...)
}

func printVersionInfo() error {
	fmt.Fprintf(os.Stdout, "%s:\n", themis.ApplicationName)
	fmt.Fprintf(os.Stdout, "  version: \t%s\n", themis.Version)
	fmt.Fprintf(os.Stdout, "  go version: \t%s\n", runtime.Version())
	fmt.Fprintf(os.Stdout, "  built time: \t%s\n", themis.BuildTime)
	fmt.Fprintf(os.Stdout, "  git commit: \t%s\n", themis.GitCommit)
	fmt.Fprintf(os.Stdout, "  os/arch: \t%s/%s\n", runtime.GOOS, runtime.GOARCH)

	return pflag.ErrHelp
}
