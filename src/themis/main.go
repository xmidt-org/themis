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
	"config"
	"fmt"
	"key"
	"os"
	"random"
	"strings"
	"token"
	"xhealth"
	"xhttp/xhttpserver"
	"xlog"
	"xlog/xloghttp"

	"github.com/go-kit/kit/log"

	"github.com/spf13/pflag"
	"go.uber.org/fx"
)

const (
	applicationName    = "themis"
	applicationVersion = "0.0.0"
)

func initialize(e config.Environment) error {
	var (
		file = e.FlagSet.StringP("file", "f", "", "the configuration file to use.  Overrides the search path.")
		dev  = e.FlagSet.Bool("dev", false, "development mode")
		iss  = e.FlagSet.String("iss", "", "the name of the issuer to put into claims.  Overrides configuration.")
	)

	e.FlagSet.BoolP("enable-app-logging", "e", false, "enables logging output from the uber/fx App")

	err := e.FlagSet.Parse(e.Arguments)
	if err != nil {
		return err
	}

	switch {
	case *dev:
		e.Viper.SetConfigType("yaml")
		err = e.Viper.ReadConfig(strings.NewReader(devMode))

	case len(*file) > 0:
		e.Viper.SetConfigFile(*file)
		err = e.Viper.ReadInConfig()

	default:
		e.Viper.SetConfigName(e.Name)
		e.Viper.AddConfigPath(".")
		e.Viper.AddConfigPath(fmt.Sprintf("$HOME/.%s", e.Name))
		e.Viper.AddConfigPath(fmt.Sprintf("/etc/%s", e.Name))
		err = e.Viper.ReadInConfig()
	}

	if err != nil {
		return err
	}

	if len(*iss) > 0 {
		e.Viper.Set("issuer.claims.iss", *iss)
	}

	return nil
}

func createPrinter(logger log.Logger, e config.Environment) fx.Printer {
	if v, err := e.FlagSet.GetBool("enable-app-logging"); v && err == nil {
		return xlog.Printer{Logger: logger}
	}

	return config.DiscardPrinter{}
}

func main() {
	var (
		b = config.Bootstrap{
			Name:        applicationName,
			Initializer: initialize,
			Optioners: config.Optioners{
				xlog.Unmarshaller("log", createPrinter),
			},
		}

		app = fx.New(
			b.Provide(),
			provideMetrics("prometheus"),
			fx.Provide(
				xhealth.Unmarshal("health"),
				random.Provide,
				key.Provide,
				token.Unmarshal("token"),
				xloghttp.ProvideStandardBuilders,
				provideClient("client"),
			),
			fx.Invoke(
				RunKeyServer("servers.key"),
				RunIssuerServer("servers.issuer"),
				RunClaimsServer("servers.claims"),
				RunMetricsServer("servers.metrics"),
				RunHealthServer("servers.health"),
				xhttpserver.InvokeOptional("servers.pprof", xhttpserver.AddPprofRoutes),
			),
		)
	)

	if err := app.Err(); err != nil {
		if err == pflag.ErrHelp {
			return
		}

		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	app.Run()
}
