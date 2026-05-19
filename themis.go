// SPDX-FileCopyrightText: 2019 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package themis

import (
	"time"

	"github.com/InVisionApp/go-health"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xmidt-org/candlelight"
	"github.com/xmidt-org/sallust"
	"github.com/xmidt-org/themis/v2/config"
	"github.com/xmidt-org/themis/v2/key"
	"github.com/xmidt-org/themis/v2/random"
	"github.com/xmidt-org/themis/v2/token"
	"github.com/xmidt-org/themis/v2/xhealth"
	"github.com/xmidt-org/themis/v2/xhttp/xhttpclient"
	"github.com/xmidt-org/themis/v2/xhttp/xhttpserver"
	"github.com/xmidt-org/themis/v2/xmetrics/xmetricshttp"

	"go.uber.org/fx"
)

const (
	ApplicationName = "themis"
)

var (
	GitCommit = "undefined"
	Version   = "undefined"
	BuildTime = "undefined"
)

func New(opts fx.Option) (*fx.App, error) {
	app := fx.New(provideAppOptions(opts))

	return app, app.Err()
}

func provideAppOptions(opts fx.Option) fx.Option {
	return fx.Options(opts,
		fx.Module(ApplicationName,
			sallust.WithLogger(),
			provideMetrics(),
			token.ProvideMetrics(),
			fx.Provide(
				config.ProvideViper,
				func(u config.Unmarshaller) (c sallust.Config, err error) {
					err = u.UnmarshalKey("log", &c)
					return
				},
				xhealth.Unmarshal("health"),
				random.Provide,
				key.Provide,
				token.Unmarshal("token"),
				token.RemoteClaimsEndpoint,
				token.TokenFactory(),
				xmetricshttp.Unmarshal("prometheus", promhttp.HandlerOpts{}),
				provideServerChainFactory,
				xhttpclient.Unmarshal{Key: "client"}.Provide,
				xhttpserver.Unmarshal{Key: "servers.key", Optional: true}.Annotated(),
				xhttpserver.Unmarshal{Key: "servers.issuer", Optional: true}.Annotated(),
				xhttpserver.Unmarshal{Key: "servers.claims", Optional: true}.Annotated(),
				xhttpserver.Unmarshal{Key: "servers.metrics", Optional: true}.Annotated(),
				xhttpserver.Unmarshal{Key: "servers.health", Optional: true}.Annotated(),
				xhttpserver.Unmarshal{Key: "servers.pprof", Optional: true}.Annotated(),
				candlelight.New,
				func(u config.Unmarshaller) (candlelight.Config, error) {
					var config candlelight.Config
					err := u.UnmarshalKey("tracing", &config)
					config.ApplicationName = ApplicationName
					return config, err
				},
				fx.Private,
			),
			fx.Invoke(
				xhealth.ApplyChecks(
					&health.Config{
						Name:     ApplicationName,
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
				BuildPprofRoutes,
				CheckServerRequirements,
			),
		))
}
