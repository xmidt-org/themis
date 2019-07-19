package main

import (
	"xmetrics"
	"xmetrics/xmetricshttp"

	"github.com/go-kit/kit/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
)

type ServerMetricsIn struct {
	fx.In

	RequestCount    metrics.Counter   `name:"server_request_count"`
	RequestDuration metrics.Histogram `name:"server_request_duration_seconds"`
	InFlight        metrics.Gauge     `name:"server_requests_in_flight"`
}

type InstrumentHandlerOut struct {
	fx.Out

	RequestCount    xmetricshttp.InstrumentHandlerCounter
	RequestDuration xmetricshttp.InstrumentHandlerDuration
	InFlight        xmetricshttp.InstrumentHandlerInFlight
}

// provideMetrics builds the various metrics components needed by the issuer
func provideMetrics() fx.Option {
	return fx.Provide(
		xmetrics.ProvideCounter(
			prometheus.CounterOpts{
				Name: "server_request_count",
				Help: "total HTTP requests made to servers",
			},
			xmetricshttp.CodeLabel,
			xmetricshttp.MethodLabel,
		),
		xmetrics.ProvideHistogram(
			prometheus.HistogramOpts{
				Name: "server_request_duration_seconds",
				Help: "tracks server request durations in seconds",
			},
			xmetricshttp.CodeLabel,
			xmetricshttp.MethodLabel,
		),
		xmetrics.ProvideGauge(
			prometheus.GaugeOpts{
				Name: "server_requests_in_flight",
				Help: "tracks the current number of server requests currently being processed",
			},
			xmetricshttp.CodeLabel,
			xmetricshttp.MethodLabel,
		),
		func(in ServerMetricsIn) InstrumentHandlerOut {
			return InstrumentHandlerOut{
				RequestCount: xmetricshttp.InstrumentHandlerCounter{
					Reporter: xmetricshttp.NewCounterReporter(in.RequestCount),
					Labeller: xmetricshttp.StandardLabeller{},
				},
				RequestDuration: xmetricshttp.InstrumentHandlerDuration{
					Reporter: xmetricshttp.NewHistogramReporter(in.RequestDuration),
					Labeller: xmetricshttp.StandardLabeller{},
				},
				InFlight: xmetricshttp.InstrumentHandlerInFlight{
					Reporter: xmetricshttp.NewGaugeReporter(in.InFlight, true),
					Labeller: xmetricshttp.StandardLabeller{},
				},
			}
		},
	)
}
