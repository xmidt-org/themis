// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package token

import (
	"encoding/json"
	"time"

	"github.com/xmidt-org/themis/key"
)

const (
	DefaultTrustLevelNoCertificates   = 0
	DefaultTrustLevelExpiredUntrusted = 100
	DefaultTrustLevelExpiredTrusted   = 1000
	DefaultTrustLevelUntrusted        = 1000
	DefaultTrustLevelTrusted          = 1000
)

// RemoteClaims describes a remote HTTP endpoint that can produce claims given the
// metadata from a token request.
type RemoteClaims struct {
	// Method is the HTTP method used to invoke the URL
	Method string

	// URL is the remote endpoint that is expected to receive Request.Metadata and return a JSON document
	// which is merged into the token claims
	URL string
}

// Value describes how to extract a key/value pair from either an HTTP request or from configuration.
type Value struct {
	// Key is the key to use for this value.  Typically, this is the name of a claim.
	Key string

	// Header is an HTTP header from which the value is pulled
	Header string

	// Parameter is a URL query parameter (including form data) from which the value is pulled
	Parameter string

	// Variable is a URL gorilla/mux variable from with the value is pulled
	Variable string

	// JSON is the value embedded as a JSON snippet.  If this field is set, Value is ignored.
	// Using this field is convenient to avoid viper's lowercasing of keys.  It's also handy
	// to embed arbitrary structures in claims.
	JSON string

	// Value is the statically assigned value from configuration
	Value interface{}
}

// IsFromHTTP tests if this value is extracted from an HTTP request
func (v Value) IsFromHTTP() bool {
	return len(v.Header) > 0 || len(v.Parameter) > 0 || len(v.Variable) > 0
}

// IsStatic tests if this value is statically configured and does not
// come from an HTTP request.
func (v Value) IsStatic() bool {
	return len(v.JSON) > 0 || v.Value != nil
}

// RawMessage precomputes the JSON  for this value.  If the JSON field is set,
// it is verified by unmarshaling.  Otherwise, the Value field is marshaled.
func (v Value) RawMessage() (json.RawMessage, error) {
	switch {
	case len(v.JSON) > 0:
		raw := []byte(v.JSON)
		var m map[string]interface{}
		err := json.Unmarshal(raw, &m)
		return json.RawMessage(raw), err

	case v.Value != nil:
		raw, err := json.Marshal(v.Value)
		return json.RawMessage(raw), err

	default:
		return json.RawMessage(nil), nil
	}
}

// PartnerID describes how to extract the partner id from an HTTP request.  Partner IDs
// require some special processing.
type PartnerID struct {
	// Claim is the name of the claim key for the partner id.  If unset, no claim is set.
	Claim string

	// Metadata is the name of the metadata key for the partner id.  If unset, no metadata
	// is set and thus the partner id won't be transmitted to remote systems.
	Metadata string

	// Header is the HTTP header containing the partner id
	Header string

	// Parameter is the HTTP parameter containing the partner id
	Parameter string

	// Default is the default value for the partner id
	Default string
}

// Trust describes the various levels of trust based upon client
// certificate state.
type Trust struct {
	// NoCertificates is the trust level to set when no client certificates are present.
	// If unset, DefaultTrustLevelNoCertificates is used.
	NoCertificates int

	// ExpiredUntrusted is the trust level to set when a certificate has both expired
	// and is within an CA chain that we do not trust.
	//
	// If unset, DefaultTrustLevelExpiredTrusted is used.
	ExpiredUntrusted int

	// ExpiredTrusted is the trust level to set when a certificate has both expired
	// and IS within a trusted CA chain.
	//
	// If unset, DefaultTrustLevelExpiredTrusted is used.
	ExpiredTrusted int

	// Untrusted is the trust level to set when a client has an otherwise valid
	// certificate, but that certificate is part of an untrusted chain.
	//
	// If unset, DefaultTrustLevelUntrusted is used.
	Untrusted int

	// Trusted is the trust level to set when a client certificate is part of
	//
	// If unset, DefaultTrustLevelTrusted is used.
	// a trusted CA chain.
	Trusted int
}

// enforceDefaults returns a Trust that has ensures any unset values are
// set to their defaults.
func (t Trust) enforceDefaults() (other Trust) {
	other = t
	if other.NoCertificates <= 0 {
		other.NoCertificates = DefaultTrustLevelNoCertificates
	}

	if other.ExpiredUntrusted <= 0 {
		other.ExpiredUntrusted = DefaultTrustLevelExpiredUntrusted
	}

	if other.ExpiredTrusted <= 0 {
		other.ExpiredTrusted = DefaultTrustLevelExpiredTrusted
	}

	if other.Untrusted <= 0 {
		other.Untrusted = DefaultTrustLevelUntrusted
	}

	if other.Trusted <= 0 {
		other.Trusted = DefaultTrustLevelTrusted
	}

	return
}

// ClientCertificates describes how peer certificates are to be handled when
// it comes to issuing tokens.
type ClientCertificates struct {
	// RootCAFile is the PEM bundle of certificates used for client certificate verification.
	// If unset, the system verifier and/or bundle is used.
	//
	// Generally, this value should be the same as the the mtls.clientCACertificateFile.
	RootCAFile string

	// IntermediatesFile is the PEM bundle of certificates used for client certificate verification.
	// If unset, no intermediary certificates are considered.
	IntermediatesFile string

	// Trust defines the trust levels to set for various situations involving
	// client certificates.
	Trust Trust
}

// Options holds the configurable information for a token Factory
type Options struct {
	// Alg is the required JWT signing algorithm to use
	Alg string

	// ClientCertificates describes how peer certificates affect the issued tokens.
	// If unset, client certificates are not considered when issuing tokens.
	ClientCertificates *ClientCertificates

	// Key describes the signing key to use
	Key key.Descriptor

	// Claims is an optional map of claims to add to every token emitted by this factory.
	// Any claims here can be overridden by claims within a token Request.
	//
	// None of these claims receive any special processing.  They are copied as is from the HTTP request
	// or statically from configuration.  For special processing around the partner id, set the PartnerID field.
	Claims []Value

	// Metadata describes non-claim data, which can be statically configured or supplied via a request
	Metadata []Value

	// PartnerID is the optional partner id configuration.  If unset, no partner id processing is
	// performed, though a partner id may still be configured as part of the claims.
	PartnerID *PartnerID

	// Nonce indicates whether a nonce (jti) should be applied to each token emitted
	// by this factory.
	Nonce bool

	// DisableTime completely disables all time-based claims, such as iat.  Setting this to true
	// also causes Duration and NotBeforeDelta to be ignored.
	DisableTime bool

	// Duration specifies how long the token should be valid for.  An exp claim is set
	// using this duration from the current time if this field is positive.
	Duration time.Duration

	// DisableNotBefore specifically controls the nbf claim.
	DisableNotBefore bool

	// NotBeforeDelta is a golang duration that determines the nbf field.  This field
	// is parsed and added to the current time at the moment a token is issued.  The result
	// is set as an nbf claim.  Note that the duration may be zero or negative.
	//
	// If either DisableTime or DisableNotBefore are true, this field is ignored and no nbf claim is emitted.
	NotBeforeDelta time.Duration

	// Remote specifies an optional external system that takes metadata from a token request
	// and returns a set of claims to be merged into tokens returned by the Factory.  Returned
	// claims from the remote system do not override claims configured on the Factory.
	Remote *RemoteClaims
}
