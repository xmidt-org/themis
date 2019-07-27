package token

import (
	"key"
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

// Value represents information pulled from either the HTTP request or statically, via config.
type Value struct {
	// Header is an HTTP header from which the value is pulled
	Header string

	// Parameter is a URL query parameter (including form data) from which the value is pulled
	Parameter string

	// Variable is a URL gorilla/mux variable from with the value is pulled
	Variable string

	// Required indicates that this value is required.  Only applies to HTTP values.
	Required bool

	// Value is the statically assigned value from configuration
	Value interface{}
}

// Options holds the configurable information for a token Factory
type Options struct {
	// Alg is the required JWT signing algorithm to use
	Alg string

	// Key describes the signing key to use
	Key key.Descriptor

	// Claims is an optional map of claims to add to every token emitted by this factory.
	// Any claims here can be overridden by claims within a token Request.
	Claims map[string]Value

	// Metadata describes non-claim data, which can be statically configured or supplied via a request
	Metadata map[string]Value

	// Nonce indicates whether a nonce (jti) should be applied to each token emitted
	// by this factory.
	Nonce bool

	// DisableTime completely disables all time-based claims, such as iat.  Setting this to true
	// causes Duration and NotBeforeDelta to be ignored.
	DisableTime bool

	// Duration specifies how long the token should be valid for.  An exp claim is set
	// using this duration from the current time if this field is positive.
	Duration string

	// NotBeforeDelta is a golang duration that determines the nbf field.  If set, this field
	// is parsed and added to the current time at the moment a token is issued.  The result
	// is set as an nbf claim.  Note that the duration may be zero or negative.
	//
	// If unset, then now nbf claim is issued.
	NotBeforeDelta string

	// Remote specifies an optional external system that takes metadata from a token request
	// and returns a set of claims to be merged into tokens returned by the Factory.  Returned
	// claims from the remote system do not override claims configured on the Factory.
	Remote *RemoteClaims
}
