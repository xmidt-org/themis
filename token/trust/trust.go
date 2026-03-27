// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package trust

import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

var (
	errTrustTypeInvalid = errors.New("invalid trust type")
)

type Type int

const (
	UnknownType Type = iota
	NoCertificates
	ExpiredUntrusted
	ExpiredTrusted
	Untrusted
	Trusted
	lastType
)

var (
	trustTypeUnmarshal = map[string]Type{
		"unknown":           UnknownType,
		"no_certificates":   NoCertificates,
		"expired_untrusted": ExpiredUntrusted,
		"expired_trusted":   ExpiredTrusted,
		"untrusted":         Untrusted,
		"trusted":           Trusted,
	}
	trustTypeMarshal = map[Type]string{
		UnknownType:      "unknown",
		NoCertificates:   "no_certificates",
		ExpiredUntrusted: "expired_untrusted",
		ExpiredTrusted:   "expired_trusted",
		Untrusted:        "untrusted",
		Trusted:          "trusted",
	}
)

// String returns a human-readable string representation for an existing Type,
// otherwise String returns the unknownEnum string value.
func (t Type) String() string {
	if value, ok := trustTypeMarshal[t]; ok {
		return value
	}

	return "unknown_trust_type"
}

// UnmarshalText returns the Type's enum value
func (t *Type) UnmarshalText(b []byte) error {
	s := string(b)
	r, ok := trustTypeUnmarshal[s]
	if !ok {
		return fmt.Errorf("%w: '%s' does not match any valid options: %v",
			errTrustTypeInvalid, s, slices.Sorted(maps.Keys(trustTypeUnmarshal)))
	}

	*t = r
	return nil
}

// MarshalText returns the Type's string value
func (t Type) MarshalText() ([]byte, error) {
	if t < UnknownType || t >= lastType {
		return []byte{}, errTrustTypeInvalid
	}

	return []byte(t.String()), nil
}
