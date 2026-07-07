// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package token

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"strings"
)

type tracingHeadersKey struct{}

type tlsDetailsKey struct{}

type tlsDetails struct {
	TLS           tls.ConnectionState
	Roots         *x509.CertPool
	Intermediates *x509.CertPool
	Trust         Trust
}

// ConnectionDetails returns everything themis knows about the device tls connection and the required details to compute its associated trust value.
func ConnectionDetails(ctx context.Context) (tls tlsDetails, ok bool) {
	tls, ok = ctx.Value(tlsDetailsKey{}).(tlsDetails)

	return
}

// SetConnectionDetails associates a tlsDetails with the given context.
func SetConnectionDetails(ctx context.Context, tls tlsDetails) context.Context {
	return context.WithValue(
		ctx,
		tlsDetailsKey{},
		tls,
	)
}

func WithTracingHeaders(ctx context.Context, r *http.Request) context.Context {
	headers := make(http.Header)
	for key, values := range r.Header {
		lowerKey := strings.ToLower(key)
		if lowerKey == "traceparent" || lowerKey == "tracestate" {
			headers[key] = values
		}
	}

	return context.WithValue(ctx, tracingHeadersKey{}, headers)
}

func TracingHeadersFromContext(ctx context.Context) http.Header {
	if headers, ok := ctx.Value(tracingHeadersKey{}).(http.Header); ok {
		return headers
	}

	return nil
}
