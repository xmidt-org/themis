package key

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xmidt-org/themis/random"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestProvide(t *testing.T) {
	t.Run("WithCustomRandom", func(t *testing.T) {
		var (
			assert = assert.New(t)

			registry Registry
			handler  Handler

			app = fxtest.New(
				t,
				fx.Provide(
					random.Provide,
					Provide,
				),
				fx.Populate(&registry, &handler),
			)
		)

		app.RequireStart()
		assert.NotNil(registry)
		assert.NotNil(handler)

		app.RequireStop()
	})

	t.Run("WithDefaultRandom", func(t *testing.T) {
		var (
			assert = assert.New(t)

			registry Registry
			handler  Handler

			app = fxtest.New(
				t,
				fx.Provide(
					Provide,
				),
				fx.Populate(&registry, &handler),
			)
		)

		app.RequireStart()
		assert.NotNil(registry)
		assert.NotNil(handler)

		app.RequireStop()
	})
}
