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
	"key"
	"os"
	"random"
	"strings"
	"token"
	"xhttp"
	"xlog"
	"xlog/xloghttp"
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

func initViper(name string, arguments []string) (*pflag.FlagSet, *viper.Viper, error) {
	var (
		fs   = pflag.NewFlagSet(name, pflag.ContinueOnError)
		file = fs.StringP("file", "f", "", "the configuration file to use.  Overrides the search path.")
		dev  = fs.BoolP("dev", "", false, "development node")
		iss  = fs.StringP("iss", "", "", "the name of the issuer to put into claims.  Overrides configuration.")
	)

	if err := fs.Parse(arguments); err != nil {
		return nil, nil, err
	}

	v := viper.New()
	switch {
	case *dev:
		v.SetConfigType("yaml")
		if err := v.ReadConfig(strings.NewReader(devMode)); err != nil {
			return nil, nil, err
		}

	case len(*file) > 0:
		v.SetConfigFile(*file)
		if err := v.ReadInConfig(); err != nil {
			return nil, nil, err
		}

	default:
		v.SetConfigName(name)
		v.AddConfigPath(".")
		v.AddConfigPath(fmt.Sprintf("$HOME/.%s", name))
		v.AddConfigPath(fmt.Sprintf("/etc/%s", name))
		if err := v.ReadInConfig(); err != nil {
			return nil, nil, err
		}
	}

	if len(*iss) > 0 {
		v.Set("issuer.claims.iss", *iss)
	}

	return fs, v, nil
}

func main() {
	fs, v, err := initViper(applicationName, os.Args[1:])
	if err == pflag.ErrHelp {
		os.Exit(0)
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to initialize viper: %s", err)
		os.Exit(1)
	}

	logger, err := xlog.Unmarshal("log", v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to initialize logging: %s", err)
		os.Exit(1)
	}

	app := fx.New(
		fx.Logger(xlog.Printer{Logger: logger}),
		fx.Provide(
			func() (*pflag.FlagSet, *viper.Viper, log.Logger) {
				return fs, v, logger
			},
			random.Provide,
			key.Provide,
			token.Unmarshal("token"),
			func() []xloghttp.ParameterBuilder {
				return []xloghttp.ParameterBuilder{
					xloghttp.Header("X-Midt-Mac-Address"),
					xloghttp.Header("X-Midt-Serial-Number"),
					xloghttp.Header("X-Midt-Uuid"),
				}
			},
			xhttp.ProvideParseForm,
			xhttp.UnmarshalResponseHeaders("responseHeaders"),
			xmetrics.Unmarshal("prometheus", promhttp.HandlerOpts{}),
		),
		fx.Invoke(
			RunKeyServer("servers.key"),
			RunIssuerServer("servers.issuer"),
			RunMetricsServer("servers.metrics"),
		),
	)

	if err := app.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to start: %s", err)
		os.Exit(2)
	}

	app.Run()
}
