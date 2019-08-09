package main

import (
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

	RequestCount     *prometheus.CounterVec   `name:"server_request_count"`
	RequestDuration  *prometheus.HistogramVec `name:"server_request_duration_ms"`
	RequestsInFlight *prometheus.GaugeVec     `name:"server_requests_in_flight"`
}

type ClientInstrumentsIn struct {
	fx.In

	RequestCount     xmetricshttp.RoundTripperCounter  `name:"client_request_count"`
	RequestDuration  xmetricshttp.RoundTripperDuration `name:"client_request_duration_ms"`
	RequestsInFlight xmetricshttp.RoundTripperInFlight `name:"client_requests_in_flight"`
}

// provideMetrics builds the various metrics components needed by the issuer
func provideMetrics(configKey string) fx.Option {
	clientLabellers := xmetricshttp.NewClientLabellers(
		xmetricshttp.CodeLabeller{},
		xmetricshttp.MethodLabeller{},
	)

	return fx.Provide(
		xmetricshttp.Unmarshal(configKey, promhttp.HandlerOpts{}),
		xmetrics.ProvideCounterVec(
			prometheus.CounterOpts{
				Name: "server_request_count",
				Help: "total incoming HTTP requests",
			},
			xmetricshttp.DefaultCodeLabel,
			xmetricshttp.DefaultMethodLabel,
			ServerLabel,
		),
		xmetrics.ProvideHistogramVec(
			prometheus.HistogramOpts{
				Name: "server_request_duration_ms",
				Help: "tracks incoming request durations in ms",
			},
			xmetricshttp.DefaultCodeLabel,
			xmetricshttp.DefaultMethodLabel,
			ServerLabel,
		),
		xmetrics.ProvideGaugeVec(
			prometheus.GaugeOpts{
				Name: "server_requests_in_flight",
				Help: "tracks the current number of incoming requests being processed",
			},
			ServerLabel,
		),
		xmetricshttp.ProvideRoundTripperCounter(
			prometheus.CounterOpts{
				Name: "client_request_count",
				Help: "total outgoing HTTP requests",
			},
			clientLabellers,
		),
		xmetricshttp.ProvideRoundTripperDurationHistogram(
			prometheus.HistogramOpts{
				Name: "client_request_count",
				Help: "total outgoing HTTP requests",
			},
			clientLabellers,
		),
		xmetricshttp.ProvideRoundTripperInFlight(
			prometheus.GaugeOpts{
				Name: "client_requests_in_flight",
				Help: "tracks the current number of incoming requests being processed",
			},
		),
	)
}

// metricsMiddleware is a helper function that creates a chain of middleware for gorilla/mux given
// the common serverside metrics.  Server metrics have an extra label that client metrics don't have.
func metricsMiddleware(in ServerMetricsIn, label string) []mux.MiddlewareFunc {
	curryLabel := prometheus.Labels{
		ServerLabel: label,
	}

	serverLabellers := xmetricshttp.NewServerLabellers(
		xmetricshttp.CodeLabeller{},
		xmetricshttp.MethodLabeller{},
	)

	return []mux.MiddlewareFunc{
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
	}
}
