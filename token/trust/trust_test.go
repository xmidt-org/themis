// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package trust

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrustUnmarshalling(t *testing.T) {
	tests := []struct {
		description string
		config      []byte
		err         error
	}{
		{
			description: "UnknownType valid",
			config:      []byte("unknown"),
		},
		{
			description: "NoCertificates valid",
			config:      []byte("no_certificates"),
		},
		{
			description: "ExpiredUntrusted valid",
			config:      []byte("expired_untrusted"),
		},
		{
			description: "ExpiredTrusted valid",
			config:      []byte("expired_trusted"),
		},
		{
			description: "Untrusted valid",
			config:      []byte("untrusted"),
		},
		{
			description: "Trusted valid",
			config:      []byte("trusted"),
		},
		{
			description: "Nonexistent policy invalid",
			config:      []byte("FOOBAR"),
			err:         errTrustTypeInvalid,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			var p Type

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

func TestTrustMarshalling(t *testing.T) {
	tests := []struct {
		description string
		trustType   Type
		config      []byte
		err         error
	}{
		{
			description: "UnknownType valid",
			trustType:   UnknownType,
			config:      []byte("unknown"),
		},
		{
			description: "NoCertificates valid",
			trustType:   NoCertificates,
			config:      []byte("no_certificates"),
		},
		{
			description: "ExpiredUntrusted valid",
			trustType:   ExpiredUntrusted,
			config:      []byte("expired_untrusted"),
		},
		{
			description: "ExpiredTrusted valid",
			trustType:   ExpiredTrusted,
			config:      []byte("expired_trusted"),
		},
		{
			description: "Untrusted valid",
			trustType:   Untrusted,
			config:      []byte("untrusted"),
		},
		{
			description: "Trusted valid",
			trustType:   Trusted,
			config:      []byte("trusted"),
		},
		{
			description: "Nonexistent policy invalid",
			trustType:   lastType,
			config:      []byte("FOOBAR"),
			err:         errTrustTypeInvalid,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			assert.NotEmpty(tc.trustType.String())
			b, err := tc.trustType.MarshalText()
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
