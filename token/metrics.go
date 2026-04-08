// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package token

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/xmidt-org/themis/xmetrics"
	"go.uber.org/fx"
)

// Metric names.
const (
	RemoteClaimsAPIResultCounter            = "remote_claims_api_result_total"
	RemoteClaimsAPIRequestDurationHistogram = "remote_claims_api_request_duration_seconds"
)

// Metric label keys for API Result counter.
const (
	EndpointLabelKey = "endpoint"
	MethodLabelKey   = "method"
	CodeLabelKey     = "code"
	OutcomeLabelKey  = "outcome"
)

// Metric label values for method.
const (
	GetMethod = "get"
)

// Metric label values for outcomes.
const (
	FailOutcome    = "fail"
	SuccessOutcome = "success"
	UnknownOutcome = "unknown"
)

// ProvideMetrics returns the Metrics for the App.
func ProvideMetrics() fx.Option {
	return fx.Provide(
		xmetrics.ProvideCounterVec(
			prometheus.CounterOpts{
				Name: RemoteClaimsAPIResultCounter,
				Help: "The total number of requests to the remote claims API.",
			},
			EndpointLabelKey,
			MethodLabelKey,
			CodeLabelKey,
			OutcomeLabelKey,
		),
		xmetrics.ProvideHistogramVec(
			prometheus.HistogramOpts{
				Name: RemoteClaimsAPIRequestDurationHistogram,
				Help: "a histogram of latencies for requests on the remote claims API",
			},
			EndpointLabelKey,
			MethodLabelKey,
			CodeLabelKey,
		),
	)
}
