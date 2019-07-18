package main

import (
	"key"
	"token"
	"xhealth"
	"xhttp"
	"xhttp/xhttpserver"
	"xlog/xloghttp"
	"xmetrics/xmetricshttp"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"go.uber.org/fx"
)

type CommonIn struct {
	fx.In

	ParameterBuilders []xloghttp.ParameterBuilder `optional:"true"`
	ResponseHeaders   xhttp.ResponseHeaders
}

type KeyServerIn struct {
	xhttpserver.ServerIn
	CommonIn

	Handler key.Handler
}

func RunKeyServer(serverConfigKey string) func(KeyServerIn) error {
	return func(in KeyServerIn) error {
		return xhttpserver.Run(
			serverConfigKey,
			in.ServerIn,
			func(router *mux.Router, l log.Logger) error {
				router.Handle("/key/{kid}", in.Handler).Methods("GET")
				router.Use(
					xloghttp.Logging{Base: l, Builders: in.ParameterBuilders}.Then,
					in.ResponseHeaders.Then,
				)

				return nil
			},
		)
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
		return xhttpserver.Optional(
			xhttpserver.Run(
				serverConfigKey,
				in.ServerIn,
				func(router *mux.Router, l log.Logger) error {
					router.Handle("/issue", in.IssueHandler).Methods("GET")
					router.Use(
						xloghttp.Logging{Base: l, Builders: in.ParameterBuilders}.Then,
						in.ParseForm.Then,
						in.ResponseHeaders.Then,
					)

					return nil
				},
			),
		)
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
		return xhttpserver.Optional(
			xhttpserver.Run(
				serverConfigKey,
				in.ServerIn,
				func(router *mux.Router, l log.Logger) error {
					router.Handle("/claims", in.ClaimsHandler).Methods("GET")
					router.Use(
						xloghttp.Logging{Base: l, Builders: in.ParameterBuilders}.Then,
						in.ParseForm.Then,
						in.ResponseHeaders.Then,
					)

					return nil
				},
			),
		)
	}
}

type MetricsServerIn struct {
	xhttpserver.ServerIn
	CommonIn

	Handler xmetricshttp.Handler
}

func RunMetricsServer(serverConfigKey string) func(MetricsServerIn) error {
	return func(in MetricsServerIn) error {
		return xhttpserver.Run(
			serverConfigKey,
			in.ServerIn,
			func(router *mux.Router, l log.Logger) error {
				router.Handle("/metrics", in.Handler).Methods("GET")
				router.Use(
					xloghttp.Logging{Base: l, Builders: in.ParameterBuilders}.Then,
					in.ResponseHeaders.Then,
				)

				return nil
			},
		)
	}
}

type HealthServerIn struct {
	xhttpserver.ServerIn
	CommonIn

	Handler xhealth.Handler
}

func RunHealthServer(serverConfigKey string) func(HealthServerIn) error {
	return func(in HealthServerIn) error {
		return xhttpserver.Run(
			serverConfigKey,
			in.ServerIn,
			func(router *mux.Router, l log.Logger) error {
				router.Handle("/health", in.Handler).Methods("GET")
				router.Use(
					xloghttp.Logging{Base: l, Builders: in.ParameterBuilders}.Then,
					in.ResponseHeaders.Then,
				)

				return nil
			},
		)
	}
}
