// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0 
package random

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestProvide(t *testing.T) {
	var (
		assert = assert.New(t)

		randomness io.Reader
		noncer     Noncer

		app = fxtest.New(
			t,
			fx.Provide(Provide),
			fx.Populate(&randomness),
			fx.Populate(&noncer),
		)
	)

	app.RequireStart()
	assert.NotNil(randomness)
	assert.NotNil(noncer)
	app.RequireStop()
}
