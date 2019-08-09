package main

import (
	"key"
	"token"
	"xhealth"
	"xmetrics/xmetricshttp"

	"github.com/gorilla/mux"
	"go.uber.org/fx"
)

type KeyRoutesIn struct {
	fx.In
	ServerMetricsIn
	Router  *mux.Router `name:"servers.key"`
	Handler key.Handler
}

func BuildKeyRoutes(in KeyRoutesIn) {
	in.Router.Handle("/key/{kid}", in.Handler).Methods("GET")
	in.Router.Use(metricsMiddleware(in.ServerMetricsIn, "key")...)
}

type IssuerRoutesIn struct {
	fx.In
	ServerMetricsIn
	Router  *mux.Router `name:"servers.issuer"`
	Handler token.IssueHandler
}

func BuildIssuerRoutes(in IssuerRoutesIn) {
	in.Router.Handle("/issue", in.Handler).Methods("GET")
	in.Router.Use(metricsMiddleware(in.ServerMetricsIn, "issuer")...)
}

type ClaimsRoutesIn struct {
	fx.In
	ServerMetricsIn
	Router  *mux.Router `name:"servers.claims"`
	Handler token.ClaimsHandler
}

func BuildClaimsRoutes(in ClaimsRoutesIn) {
	in.Router.Handle("/claims", in.Handler).Methods("GET")
	in.Router.Use(metricsMiddleware(in.ServerMetricsIn, "claims")...)
}

type MetricsRoutesIn struct {
	fx.In
	Router  *mux.Router `name:"servers.metrics" optional:"true"`
	Handler xmetricshttp.Handler
}

func BuildMetricsRoutes(in MetricsRoutesIn) {
	if in.Router != nil {
		in.Router.Handle("/metrics", in.Handler).Methods("GET")
	}
}

type HealthRoutesIn struct {
	fx.In
	Router  *mux.Router `name:"servers.health" optional:"true"`
	Handler xhealth.Handler
}

func BuildHealthRoutes(in HealthRoutesIn) {
	if in.Router != nil {
		in.Router.Handle("/health", in.Handler).Methods("GET")
	}
}
