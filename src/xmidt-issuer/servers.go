package main

import (
	"key"
	"token"
	"xhttp"
	"xhttp/xhttpserver"
	"xlog/xloghttp"
	"xmetrics"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"go.uber.org/fx"
)

type CommonIn struct {
	fx.In

	LoggerParameterBuilders []xloghttp.ParameterBuilder `optional:"true"`
	ResponseHeaders         xhttp.ResponseHeaders
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
				router.Handle("/key/{kid}", in.Handler)
				router.Use(
					xloghttp.Logging{Base: l, Builders: in.LoggerParameterBuilders}.Then,
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

	ParseForm xhttp.ParseForm
	Handler   token.Handler
}

func RunIssuerServer(serverConfigKey string) func(IssuerServerIn) error {
	return func(in IssuerServerIn) error {
		return xhttpserver.Run(
			serverConfigKey,
			in.ServerIn,
			func(router *mux.Router, l log.Logger) error {
				router.Handle("/issue", in.Handler)
				router.Use(
					xloghttp.Logging{Base: l, Builders: in.LoggerParameterBuilders}.Then,
					in.ParseForm.Then,
					in.ResponseHeaders.Then,
				)

				return nil
			},
		)
	}
}

type MetricsServerIn struct {
	xhttpserver.ServerIn
	CommonIn

	Handler xmetrics.Handler
}

func RunMetricsServer(serverConfigKey string) func(MetricsServerIn) error {
	return func(in MetricsServerIn) error {
		return xhttpserver.Run(
			serverConfigKey,
			in.ServerIn,
			func(router *mux.Router, l log.Logger) error {
				router.Handle("/metrics", in.Handler)
				router.Use(
					xloghttp.Logging{Base: l, Builders: in.LoggerParameterBuilders}.Then,
					in.ResponseHeaders.Then,
				)

				return nil
			},
		)
	}
}
