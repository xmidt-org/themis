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

	RequestCount     *prometheus.CounterVec   `name:"server_request_count"`
	RequestDuration  *prometheus.HistogramVec `name:"server_request_duration_ms"`
	RequestsInFlight *prometheus.GaugeVec     `name:"server_requests_in_flight"`
}

func provideServerChainFactory(in ServerChainIn) xhttpserver.ChainFactory {
	return xhttpserver.ChainFactoryFunc(func(o xhttpserver.Options) (alice.Chain, error) {
		var (
			curryLabel = prometheus.Labels{
				ServerLabel: o.Name,
			}

			serverLabellers = xmetricshttp.NewServerLabellers(
				xmetricshttp.CodeLabeller{},
				xmetricshttp.MethodLabeller{},
			)
		)

		requestCount, err := in.RequestCount.CurryWith(curryLabel)
		if err != nil {
			return alice.Chain{}, err
		}

		requestDuration, err := in.RequestDuration.CurryWith(curryLabel)
		if err != nil {
			return alice.Chain{}, err
		}

		requestsInFlight, err := in.RequestsInFlight.CurryWith(curryLabel)
		if err != nil {
			return alice.Chain{}, err
		}

		return alice.New(
			xmetricshttp.HandlerCounter{
				Metric:   xmetrics.LabelledCounterVec{CounterVec: requestCount},
				Labeller: serverLabellers,
			}.Then,
			xmetricshttp.HandlerDuration{
				Metric:   xmetrics.LabelledObserverVec{ObserverVec: requestDuration},
				Labeller: serverLabellers,
			}.Then,
			xmetricshttp.HandlerInFlight{
				Metric: xmetrics.LabelledGaugeVec{GaugeVec: requestsInFlight},
			}.Then,
		), nil
	})
}

type KeyRoutesIn struct {
	fx.In
	Router  *mux.Router `name:"servers.key" optional:"true"`
	Handler key.Handler
}

func BuildKeyRoutes(in KeyRoutesIn) {
	if in.Router != nil {
		in.Router.Handle("/key/{kid}", in.Handler).Methods("GET")
	}
}

type IssuerRoutesIn struct {
	fx.In
	Router  *mux.Router `name:"servers.issuer" optional:"true"`
	Handler token.IssueHandler
}

func BuildIssuerRoutes(in IssuerRoutesIn) {
	if in.Router != nil {
		in.Router.Handle("/issue", in.Handler).Methods("GET")
	}
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
