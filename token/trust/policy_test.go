// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package trust

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPolicyUnmarshalling(t *testing.T) {
	tests := []struct {
		description string
		config      []byte
		err         error
	}{
		{
			description: "UnknownPolicy valid",
			config:      []byte("unknown"),
		},
		{
			description: "Lowest valid",
			config:      []byte("lowest"),
		},
		{
			description: "Highest valid",
			config:      []byte("highest"),
		},
		{
			description: "Nonexistent policy invalid",
			config:      []byte("FOOBAR"),
			err:         errPolicyInvalid,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			var p Policy

			err := p.UnmarshalText(tc.config)
			if tc.err == nil {
				assert.NoError(err)
				assert.Equal(string(tc.config), p.String())
			} else {
				assert.Error(err)
			}
		})
	}
}

func TestPolicyMarshalling(t *testing.T) {
	tests := []struct {
		description string
		policy      Policy
		config      []byte
		err         error
	}{
		{
			description: "UnknownPolicy valid",
			policy:      UnknownPolicy,
			config:      []byte("unknown"),
		},
		{
			description: "Lowest valid",
			policy:      Lowest,
			config:      []byte("lowest"),
		},
		{
			description: "Highest valid",
			policy:      Highest,
			config:      []byte("highest"),
		},
		{
			description: "Nonexistent policy invalid",
			policy:      lastPolicy,
			config:      []byte("FOOBAR"),
			err:         errPolicyInvalid,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			assert.NotEmpty(tc.policy.String())
			b, err := tc.policy.MarshalText()
			if tc.err == nil {
				assert.NoError(err)
				assert.Equal(tc.config, b)
			} else {
				assert.Empty(b)
				assert.Error(err)
			}
		})
	}
}
