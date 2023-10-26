// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package random

import (
	"crypto/rand"
	"encoding/base64"
	"io"

	"go.uber.org/fx"
)

// RandomOut describes the components emitted by this package
type RandomOut struct {
	fx.Out

	// Random is the source of randomness to use
	Random io.Reader

	// Noncer is the nonce generation strategy
	Noncer Noncer
}

// Provide is an uber/fx provider that emits the components exposed by this package
func Provide() RandomOut {
	return RandomOut{
		Random: rand.Reader,
		Noncer: NewBase64Noncer(rand.Reader, 16, base64.RawURLEncoding),
	}
}
