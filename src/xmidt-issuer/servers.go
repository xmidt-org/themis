package main

import (
	"key"
	"token"
	"xhealth"
	"xhttp"
	"xhttp/xhttpserver"
	"xlog/xloghttp"
	"xmetrics/xmetricshttp"

	"go.uber.org/fx"
)

type CommonIn struct {
	fx.In
	ServerMetricsIn

	ParseForm         xhttp.ParseForm
	ParameterBuilders []xloghttp.ParameterBuilder `optional:"true"`
}

type KeyServerIn struct {
	xhttpserver.ServerIn
	CommonIn

	Handler key.Handler
}

func RunKeyServer(serverConfigKey string) func(KeyServerIn) error {
	return func(in KeyServerIn) error {
		_, err := xhttpserver.Run(
			serverConfigKey,
			in.ServerIn,
			func(ur xhttpserver.UnmarshalResult) error {
				ur.Router.Handle("/key/{kid}", in.Handler).Methods("GET")
				ur.Router.Use(xhttpserver.TrackWriter)
				ur.Router.Use(xloghttp.Logging{Base: ur.Logger, Builders: in.ParameterBuilders}.Then)
				ur.Router.Use(metricsMiddleware(in.ServerMetricsIn, ur)...)

				return nil
			},
		)

		return err
	}
}

type IssuerServerIn struct {
	xhttpserver.ServerIn
	CommonIn

	ParseForm    xhttp.ParseForm
	IssueHandler token.IssueHandler
}

func RunIssuerServer(serverConfigKey string) func(IssuerServerIn) error {
	return func(in IssuerServerIn) error {
		_, err := xhttpserver.Run(
			serverConfigKey,
			in.ServerIn,
			func(ur xhttpserver.UnmarshalResult) error {
				ur.Router.Handle("/issue", in.IssueHandler).Methods("GET")
				ur.Router.Use(xhttpserver.TrackWriter)
				ur.Router.Use(xloghttp.Logging{Base: ur.Logger, Builders: in.ParameterBuilders}.Then)
				ur.Router.Use(metricsMiddleware(in.ServerMetricsIn, ur)...)
				ur.Router.Use(in.ParseForm.Then)

				return nil
			},
		)

		return err
	}
}

type ClaimsServerIn struct {
	xhttpserver.ServerIn
	CommonIn

	ParseForm     xhttp.ParseForm
	ClaimsHandler token.ClaimsHandler
}

func RunClaimsServer(serverConfigKey string) func(ClaimsServerIn) error {
	return func(in ClaimsServerIn) error {
		_, err := xhttpserver.Run(
			serverConfigKey,
			in.ServerIn,
			func(ur xhttpserver.UnmarshalResult) error {
				ur.Router.Handle("/claims", in.ClaimsHandler).Methods("GET")
				ur.Router.Use(xhttpserver.TrackWriter)
				ur.Router.Use(xloghttp.Logging{Base: ur.Logger, Builders: in.ParameterBuilders}.Then)
				ur.Router.Use(metricsMiddleware(in.ServerMetricsIn, ur)...)
				ur.Router.Use(in.ParseForm.Then)

				return nil
			},
		)

		return err
	}
}

type MetricsServerIn struct {
	xhttpserver.ServerIn
	CommonIn

	ParameterBuilders []xloghttp.ParameterBuilder `optional:"true"`
	ResponseHeaders   xhttp.ResponseHeaders
	Handler           xmetricshttp.Handler
}

func RunMetricsServer(serverConfigKey string) func(MetricsServerIn) error {
	return func(in MetricsServerIn) error {
		_, err := xhttpserver.Run(
			serverConfigKey,
			in.ServerIn,
			func(ur xhttpserver.UnmarshalResult) error {
				ur.Router.Handle("/metrics", in.Handler).Methods("GET")
				ur.Router.Use(xloghttp.Logging{Base: ur.Logger, Builders: in.ParameterBuilders}.Then)

				return nil
			},
		)

		return err
	}
}

type HealthServerIn struct {
	xhttpserver.ServerIn
	CommonIn

	Handler xhealth.Handler
}

func RunHealthServer(serverConfigKey string) func(HealthServerIn) error {
	return func(in HealthServerIn) error {
		_, err := xhttpserver.Run(
			serverConfigKey,
			in.ServerIn,
			func(ur xhttpserver.UnmarshalResult) error {
				ur.Router.Handle("/health", in.Handler).Methods("GET")
				ur.Router.Use(xloghttp.Logging{Base: ur.Logger, Builders: in.ParameterBuilders}.Then)

				return nil
			},
		)

		return err
	}
}
