package main

import (
	"xhttp/xhttpserver"
	"xmetrics"
	"xmetrics/xmetricshttp"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/fx"
)

// ServerLabel is the metric label for which internal server (key, claims, etc) a metric is for
const ServerLabel = "server"

type ServerMetricsIn struct {
	fx.In

	RequestCount    *prometheus.CounterVec   `name:"server_request_count"`
	RequestDuration *prometheus.HistogramVec `name:"server_request_duration_ms"`
	InFlight        *prometheus.GaugeVec     `name:"server_requests_in_flight"`
}

// metricsMiddleware is a helper function that creates a chain of middleware for gorilla/mux given
// the common serverside metrics
func metricsMiddleware(in ServerMetricsIn, ur xhttpserver.UnmarshalResult) []mux.MiddlewareFunc {
	curryLabel := prometheus.Labels{
		ServerLabel: ur.Name,
	}

	return []mux.MiddlewareFunc{
		xmetricshttp.InstrumentHandlerCounter{
			Metric: xmetrics.LabelledCounterVec{CounterVec: in.RequestCount.MustCurryWith(curryLabel)},
			Labeller: xmetricshttp.ServerLabellers{
				xmetricshttp.CodeLabeller{},
				xmetricshttp.MethodLabeller{},
			},
		}.Then,
		xmetricshttp.InstrumentHandlerDuration{
			Metric: xmetrics.LabelledObserverVec{ObserverVec: in.RequestDuration.MustCurryWith(curryLabel)},
			Labeller: xmetricshttp.ServerLabellers{
				xmetricshttp.CodeLabeller{},
				xmetricshttp.MethodLabeller{},
			},
		}.Then,
		xmetricshttp.InstrumentHandlerInFlight{
			Metric: xmetrics.LabelledGaugeVec{GaugeVec: in.InFlight.MustCurryWith(curryLabel)},
		}.Then,
	}
}

// provideMetrics builds the various metrics components needed by the issuer
func provideMetrics() fx.Option {
	return fx.Provide(
		xmetricshttp.Unmarshal("prometheus", promhttp.HandlerOpts{}),
		xmetrics.ProvideCounterVec(
			prometheus.CounterOpts{
				Name: "server_request_count",
				Help: "total HTTP requests made to servers",
			},
			xmetricshttp.DefaultCodeLabel,
			xmetricshttp.DefaultMethodLabel,
			ServerLabel,
		),
		xmetrics.ProvideHistogramVec(
			prometheus.HistogramOpts{
				Name: "server_request_duration_ms",
				Help: "tracks server request durations in milliseconds",
			},
			xmetricshttp.DefaultCodeLabel,
			xmetricshttp.DefaultMethodLabel,
			ServerLabel,
		),
		xmetrics.ProvideGaugeVec(
			prometheus.GaugeOpts{
				Name: "server_requests_in_flight",
				Help: "tracks the current number of server requests currently being processed",
			},
			ServerLabel,
		),
	)
}
