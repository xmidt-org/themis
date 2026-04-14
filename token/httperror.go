// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package token

import (
	"context"
	"errors"
	"net"
	"net/url"
	"strings"
)

func GetRemoteClaimsReasonFromError(err error) string {
	var d *net.DNSError
	if err == nil {
		return NoErrReason
	} else if errors.Is(err, context.DeadlineExceeded) {
		// Handle as successful 2XX response from remote claims endpoint.
		return RemoteClaimsResponseNon2XXErrOkReason
	} else if errors.Is(err, context.Canceled) {
		// Handle as successful 2XX response from remote claims endpoint.
		return RemoteClaimsResponseNon2XXErrOkReason
	} else if errors.Is(err, &net.AddrError{}) {
		return AddressErrReason
	} else if errors.Is(err, &net.ParseError{}) {
		return ParseAddrErrReason
	} else if errors.Is(err, net.InvalidAddrError("")) {
		return InvalidAddrReason
	} else if errors.As(err, &d) {
		if d.IsNotFound {
			return HostNotFoundReason
		}
		return DNSErrReason
	} else if errors.Is(err, net.ErrClosed) {
		return ConnClosedReason
	} else if errors.Is(err, &net.OpError{}) {
		return OpErrReason
	} else if errors.Is(err, net.UnknownNetworkError("")) {
		return NetworkErrReason
	}

	// nolint: errorlint
	if err, ok := err.(*url.Error); ok {
		if strings.TrimSpace(strings.ToLower(err.Unwrap().Error())) == "eof" {
			return ConnectionUnexpectedlyClosedEOFReason
		}
	}

	// Custom errors.
	if errors.Is(err, ErrRemoteClaimsResponseDecodingFailure) {
		return RemoteClaimsResponseDecodingErrReason
	} else if errors.Is(err, ErrRemoteClaimsRequestEncodingFailure) {
		return RemoteClaimsRequestEncodingErrReason
	}

	var respErr RemoteClaimsResponseError
	if errors.As(err, &respErr) {
		// On their own, these are handle as successful 2XX response from remote claims endpoint.
		return RemoteClaimsResponseNon2XXErrOkReason
	}

	return UnknownReason
}
