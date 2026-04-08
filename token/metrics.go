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
	ReasonLabelKey   = "reason"
)

// Metric label values for outcomes.
const (
	FailOutcome    = "fail"
	SuccessOutcome = "success"
	UnknownOutcome = "unknown"
)

// Metric label values for reasons.
const (
	UnknownReason = "unknown"

	// DoErr failure reasons.
	DeadlineExceededReason                = "context_deadline_exceeded"
	ContextCanceledReason                 = "context_canceled"
	AddressErrReason                      = "address_error"
	ParseAddrErrReason                    = "parse_address_error"
	InvalidAddrReason                     = "invalid_address"
	DNSErrReason                          = "dns_error"
	HostNotFoundReason                    = "host_not_found"
	ConnClosedReason                      = "connection_closed"
	OpErrReason                           = "op_error"
	NetworkErrReason                      = "unknown_network_error"
	UpdateRequestURLFailedReason          = "update_request_url_failed"
	ConnectionUnexpectedlyClosedEOFReason = "connection_unexpectedly_closed_eof"
	NoErrReason                           = "no_error"

	// Custom failure reasons
	RemoteClaimsResponseDecodingErrReason = "response_decoding_error"
	RemoteClaimsRequestEncodingErrReason  = "request_encoding_error"
	RemoteClaimsResponseNon2XXErrOkReason = "non_2XX_response_is_ok"
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
			ReasonLabelKey,
		),
		xmetrics.ProvideHistogramVec(
			prometheus.HistogramOpts{
				Name: RemoteClaimsAPIRequestDurationHistogram,
				Help: "a histogram of latencies for requests on the remote claims API",
			},
			EndpointLabelKey,
			MethodLabelKey,
			CodeLabelKey,
			OutcomeLabelKey,
		),
	)
}
