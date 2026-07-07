// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package token

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/xmidt-org/themis/v2/xmetrics"
	"go.uber.org/fx"
)

// Metric names.
const (
	TrustCounter                            = "trust_total"
	RemoteClaimsAPIResultCounter            = "remote_claims_api_result_total"
	RemoteClaimsAPIRequestDurationHistogram = "remote_claims_api_request_duration_seconds"
)

// Metric label keys for API Result counter.
const (
	EndpointLabelKey  = "endpoint"
	MethodLabelKey    = "method"
	CodeLabelKey      = "code"
	OutcomeLabelKey   = "outcome"
	ReasonLabelKey    = "reason"
	TrustLabelKey     = "trust"
	IssuerCNLabelKey  = "issuer_cn"
	PartnerIDLabelKey = "partner_id"
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
	ConnectionUnexpectedlyClosedEOFReason = "connection_unexpectedly_closed_eof"
	NoErrReason                           = "no_error"

	// Trust reasons.
	NoCertificatesReason        = "no_certificates"
	ExpiredUntrustedReason      = "expired_untrusted"
	ExpiredTrustedReason        = "expired_trusted"
	UntrustedReason             = "untrusted"
	TrustedReason               = "trusted"
	UntrustedCertIssuerCNReason = "untrusted_cert_issuer_cn"

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
				Name: TrustCounter,
				Help: "The total trust.",
			},
			TrustLabelKey,
			IssuerCNLabelKey,
			PartnerIDLabelKey,
			ReasonLabelKey,
		),
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
