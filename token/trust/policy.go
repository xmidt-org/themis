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
	errPolicyInvalid = errors.New("invalid trust policy")
)

type Policy int

const (
	UnknownPolicy Policy = iota
	Lowest
	Highest
	lastPolicy
)

var (
	policyUnmarshal = map[string]Policy{
		"unknown": UnknownPolicy,
		"lowest":  Lowest,
		"highest": Highest,
	}
	policyMarshal = map[Policy]string{
		UnknownPolicy: "unknown",
		Lowest:        "lowest",
		Highest:       "highest",
	}
)

// String returns a human-readable string representation for an existing Policy,
// otherwise String returns 'unknown_trust_policy'.
func (p Policy) String() string {
	if value, ok := policyMarshal[p]; ok {
		return value
	}

	return "unknown_trust_policy"
}

// UnmarshalText returns the Policy's enum value
func (p *Policy) UnmarshalText(b []byte) error {
	s := string(b)
	r, ok := policyUnmarshal[s]
	if !ok {
		return fmt.Errorf("%w: '%s' does not match any valid options: %v",
			errPolicyInvalid, s, slices.Sorted(maps.Keys(policyUnmarshal)))
	}

	*p = r
	return nil
}

// MarshalText returns the Policy's string value
func (p Policy) MarshalText() ([]byte, error) {
	if p < UnknownPolicy || p >= lastPolicy {
		return []byte{}, errPolicyInvalid
	}

	return []byte(p.String()), nil
}
