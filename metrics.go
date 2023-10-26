// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"github.com/xmidt-org/themis/xmetrics"
	"github.com/xmidt-org/themis/xmetrics/xmetricshttp"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
)

// ServerLabel is the metric label for which internal server (key, claims, etc) a metric is for
const ServerLabel = "server"

// provideMetrics builds the application metrics and makes them available to the container
func provideMetrics() fx.Option {
	return fx.Provide(
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
		xmetrics.ProvideCounterVec(
			prometheus.CounterOpts{
				Name: "client_request_count",
				Help: "total outgoing HTTP requests",
			},
			xmetricshttp.DefaultCodeLabel,
			xmetricshttp.DefaultMethodLabel,
		),
		xmetrics.ProvideHistogramVec(
			prometheus.HistogramOpts{
				Name: "client_request_count",
				Help: "total outgoing HTTP requests",
			},
			xmetricshttp.DefaultCodeLabel,
			xmetricshttp.DefaultMethodLabel,
		),
		xmetrics.ProvideGaugeVec(
			prometheus.GaugeOpts{
				Name: "client_requests_in_flight",
				Help: "tracks the current number of incoming requests being processed",
			},
		),
	)
}
