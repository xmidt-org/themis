package main

import (
	"key"
	"token"
	"xhealth"
	"xhttp/xhttpserver"
	"xmetrics"
	"xmetrics/xmetricshttp"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
)

type ServerChainIn struct {
	fx.In
	ServerMetricsIn
}

func provideServerChainFactory(in ServerChainIn) xhttpserver.ChainFactory {
	return xhttpserver.ChainFactoryFunc(func(o xhttpserver.Options) (alice.Chain, error) {
		curryLabel := prometheus.Labels{
			ServerLabel: o.Name,
		}

		serverLabellers := xmetricshttp.NewServerLabellers(
			xmetricshttp.CodeLabeller{},
			xmetricshttp.MethodLabeller{},
		)

		return alice.New(
			xmetricshttp.HandlerCounter{
				Metric:   xmetrics.LabelledCounterVec{CounterVec: in.RequestCount.MustCurryWith(curryLabel)},
				Labeller: serverLabellers,
			}.Then,
			xmetricshttp.HandlerDuration{
				Metric:   xmetrics.LabelledObserverVec{ObserverVec: in.RequestDuration.MustCurryWith(curryLabel)},
				Labeller: serverLabellers,
			}.Then,
			xmetricshttp.HandlerInFlight{
				Metric: xmetrics.LabelledGaugeVec{GaugeVec: in.RequestsInFlight.MustCurryWith(curryLabel)},
			}.Then,
		), nil
	})
}

type KeyRoutesIn struct {
	fx.In
	Router  *mux.Router `name:"servers.key"`
	Handler key.Handler
}

func BuildKeyRoutes(in KeyRoutesIn) {
	in.Router.Handle("/key/{kid}", in.Handler).Methods("GET")
}

type IssuerRoutesIn struct {
	fx.In
	Router  *mux.Router `name:"servers.issuer"`
	Handler token.IssueHandler
}

func BuildIssuerRoutes(in IssuerRoutesIn) {
	in.Router.Handle("/issue", in.Handler).Methods("GET")
}

type ClaimsRoutesIn struct {
	fx.In
	Router  *mux.Router `name:"servers.claims"`
	Handler token.ClaimsHandler
}

func BuildClaimsRoutes(in ClaimsRoutesIn) {
	in.Router.Handle("/claims", in.Handler).Methods("GET")
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
