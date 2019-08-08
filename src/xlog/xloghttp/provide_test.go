package xloghttp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestProvideStandardBuilders(t *testing.T) {
	var (
		assert = assert.New(t)

		builders ParameterBuilders

		app = fxtest.New(t,
			fx.Provide(
				ProvideStandardBuilders,
			),
			fx.Populate(&builders),
		)
	)

	assert.NoError(app.Err())
	assert.NotEmpty(builders)
}
