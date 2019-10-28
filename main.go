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
	"os"
	"strings"
	"time"

	health "github.com/InVisionApp/go-health"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xmidt-org/themis/config"
	"github.com/xmidt-org/themis/key"
	"github.com/xmidt-org/themis/random"
	"github.com/xmidt-org/themis/token"
	"github.com/xmidt-org/themis/xhealth"
	"github.com/xmidt-org/themis/xhttp/xhttpclient"
	"github.com/xmidt-org/themis/xhttp/xhttpserver"
	"github.com/xmidt-org/themis/xlog"
	"github.com/xmidt-org/themis/xlog/xloghttp"
	"github.com/xmidt-org/themis/xmetrics/xmetricshttp"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

const (
	applicationName    = "themis"
	applicationVersion = "0.2.0"
)

func setupFlagSet(fs *pflag.FlagSet) error {
	fs.StringP("file", "f", "", "the configuration file to use.  Overrides the search path.")
	fs.Bool("dev", false, "development mode")
	fs.String("iss", "", "the name of the issuer to put into claims.  Overrides configuration.")
	fs.BoolP("debug", "d", false, "enables debug logging.  Overrides configuration.")
	return nil
}

func setupViper(in config.ViperIn, v *viper.Viper) (err error) {
	if dev, _ := in.FlagSet.GetBool("dev"); dev {
		v.SetConfigType("yaml")
		err = v.ReadConfig(strings.NewReader(devMode))
	} else if file, _ := in.FlagSet.GetString("file"); len(file) > 0 {
		v.SetConfigFile(file)
		err = v.ReadInConfig()
	} else {
		v.SetConfigName(string(in.Name))
		v.AddConfigPath(".")
		v.AddConfigPath(fmt.Sprintf("$HOME/.%s", in.Name))
		v.AddConfigPath(fmt.Sprintf("/etc/%s", in.Name))
		err = v.ReadInConfig()
	}

	if err != nil {
		return
	}

	if iss, _ := in.FlagSet.GetString("iss"); len(iss) > 0 {
		v.Set("issuer.claims.iss", iss)
	}

	if debug, _ := in.FlagSet.GetBool("debug"); debug {
		v.Set("log.level", "DEBUG")
	}

	return nil
}

func main() {
	app := fx.New(
		xlog.Logger(),
		config.CommandLine{Name: applicationName}.Provide(setupFlagSet),
		provideMetrics(),
		fx.Provide(
			config.ProvideViper(setupViper),
			xlog.Unmarshal("log"),
			xloghttp.ProvideStandardBuilders,
			xhealth.Unmarshal("health"),
			random.Provide,
			key.Provide,
			token.Unmarshal("token"),
			xmetricshttp.Unmarshal("prometheus", promhttp.HandlerOpts{}),
			provideClientChain,
			provideServerChainFactory,
			xhttpclient.Unmarshal{Key: "client"}.Provide,
			xhttpserver.Unmarshal{Key: "servers.key"}.Annotated(),
			xhttpserver.Unmarshal{Key: "servers.issuer"}.Annotated(),
			xhttpserver.Unmarshal{Key: "servers.claims"}.Annotated(),
			xhttpserver.Unmarshal{Key: "servers.metrics"}.Annotated(),
			xhttpserver.Unmarshal{Key: "servers.health"}.Annotated(),
		),
		fx.Invoke(
			xhealth.ApplyChecks(
				&health.Config{
					Name:     applicationName,
					Interval: 24 * time.Hour,
					Checker: xhealth.NopCheckable{
						Details: map[string]interface{}{
							"StartTime": time.Now().UTC().Format(time.RFC3339),
						},
					},
				},
			),
			BuildKeyRoutes,
			BuildIssuerRoutes,
			BuildClaimsRoutes,
			BuildMetricsRoutes,
			BuildHealthRoutes,
		),
	)

	switch err := app.Err(); err {
	case pflag.ErrHelp:
		return
	case nil:
		app.Run()
	default:
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
