// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package random

import (
	"crypto/rand"
	"encoding/base64"
	"io"
)

const DefaultNonceSize = 16

// Noncer is a strategy for creating nonces for JWTs, to be stored in the jti claim.
type Noncer interface {
	Nonce() (string, error)
}

type base64Noncer struct {
	random   io.Reader
	size     int
	encoding *base64.Encoding
}

func (n base64Noncer) Nonce() (string, error) {
	b := make([]byte, n.size)
	if _, err := n.random.Read(b); err != nil {
		return "", err
	}

	return n.encoding.EncodeToString(b), nil
}

// NewBase64Noncer creates a Noncer that generates a random sequence of bits encoded via
// the given base64 encoding.  All parameters have defaults:
//
// If random is nil, crypto/rand.Reader is used
// If size is nonpositive, DefaultNonceSize is used
// if encoding is nil, base64.RawURLEncoding is used
func NewBase64Noncer(random io.Reader, size int, encoding *base64.Encoding) Noncer {
	if random == nil {
		random = rand.Reader
	}

	if size <= 0 {
		size = DefaultNonceSize
	}

	if encoding == nil {
		encoding = base64.RawURLEncoding
	}

	return base64Noncer{
		random:   random,
		size:     size,
		encoding: encoding,
	}
}
