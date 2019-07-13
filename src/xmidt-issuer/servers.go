package main

import (
	"key"
	"token"
	"xhttp/xhttpserver"
	"xmetrics"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

type KeyServerIn struct {
	xhttpserver.ServerIn

	Handler key.Handler
}

func RunKeyServer(serverConfigKey string) func(KeyServerIn) error {
	return func(in KeyServerIn) error {
		return xhttpserver.Run(
			serverConfigKey,
			in.ServerIn,
			func(router *mux.Router, _ log.Logger) error {
				router.Handle("/key/{kid}", in.Handler)
				return nil
			},
		)
	}
}

type IssuerServerIn struct {
	xhttpserver.ServerIn

	Handler token.Handler
}

func RunIssuerServer(serverConfigKey string) func(IssuerServerIn) error {
	return func(in IssuerServerIn) error {
		return xhttpserver.Run(
			serverConfigKey,
			in.ServerIn,
			func(router *mux.Router, _ log.Logger) error {
				router.Handle("/issue", in.Handler)
				return nil
			},
		)
	}
}

type MetricsServerIn struct {
	xhttpserver.ServerIn

	Handler xmetrics.Handler
}

func RunMetricsServer(serverConfigKey string) func(MetricsServerIn) error {
	return func(in MetricsServerIn) error {
		return xhttpserver.Run(
			serverConfigKey,
			in.ServerIn,
			func(router *mux.Router, _ log.Logger) error {
				router.Handle("/metrics", in.Handler)
				return nil
			},
		)
	}
}
