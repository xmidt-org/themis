// Copyright 2019 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"issuer"
	"key"
	"os"
	"random"
	"token"
	"xerror"
	"xlog"
	"xmetrics"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

const (
	applicationName    = "xmidt-issuer"
	applicationVersion = "0.0.0"
)

func main() {
	var (
		fs *pflag.FlagSet
		//configFile string
		v      *viper.Viper
		logger log.Logger
	)

	err := xerror.Do(
		func() (err error) {
			fs, err = parseCommandLine(applicationName, os.Args[1:])
			return
		},
		func() (err error) {
			//configFile, err = fs.GetString(FileFlag)
			return
		},
		func() (err error) {
			//v, err = newViper(applicationName, configFile, fs)
			v = viper.New()
			return
		},
		func() (err error) {
			logger, err = xlog.NewLogger("log", v)
			return
		},
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	app := fx.New(
		fx.Logger(xlog.Printer{Logger: logger}),
		fx.Provide(
			func() (*pflag.FlagSet, *viper.Viper, log.Logger) {
				return fs, v, logger
			},
			random.Provide,
			key.Provide("servers.key"),
			token.Provide("token"),
			issuer.Provide("servers.issuer", "issuer"),
			xmetrics.Provide("servers.metrics", "prometheus", promhttp.HandlerOpts{}),
		),
		fx.Invoke(
			key.RunServer("/key/{kid}"),
			issuer.RunServer("/issue"),
			xmetrics.RunServer("/metrics"),
		),
	)

	if err := app.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to start: %s", err)
		os.Exit(2)
	}

	app.Run()
}
