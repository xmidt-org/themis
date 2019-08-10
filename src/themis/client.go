package main

import (
	"xhttp/xhttpclient"
	"xmetrics"
	"xmetrics/xmetricshttp"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
)

type ClientChainIn struct {
	fx.In
	RequestCount     *prometheus.CounterVec   `name:"client_request_count"`
	RequestDuration  *prometheus.HistogramVec `name:"client_request_duration_ms"`
	RequestsInFlight *prometheus.GaugeVec     `name:"client_requests_in_flight"`
}

// provideClientChain provides the global decoration for all HTTP clients
func provideClientChain(in ClientChainIn) xhttpclient.Chain {
	labeller := xmetricshttp.NewClientLabellers(
		xmetricshttp.CodeLabeller{},
		xmetricshttp.MethodLabeller{},
	)

	return xhttpclient.NewChain(
		xmetricshttp.RoundTripperCounter{
			Metric:   xmetrics.LabelledCounterVec{CounterVec: in.RequestCount},
			Labeller: labeller,
		}.Then,
		xmetricshttp.RoundTripperDuration{
			Metric:   xmetrics.LabelledObserverVec{ObserverVec: in.RequestDuration},
			Labeller: labeller,
		}.Then,
		xmetricshttp.RoundTripperInFlight{
			Metric: xmetrics.LabelledGaugeVec{GaugeVec: in.RequestsInFlight},
		}.Then,
	)
}
