package token

import (
	"testing"

	"github.com/xmidt-org/themis/config"
	"github.com/xmidt-org/themis/key"
	"github.com/xmidt-org/themis/xlog"

	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func testUnmarshalError(t *testing.T) {
	var (
		assert  = assert.New(t)
		factory Factory
	)

	app := fx.New(
		fx.Logger(xlog.DiscardPrinter{}),
		fx.Provide(
			config.ProvideViper(
				config.Json(`
					{
						"token": {
							"nonce": "this is not a valid bool"
						}
					}
				`),
			),
			func() key.Registry { return key.NewRegistry(nil) },
			Unmarshal("token"),
		),
		fx.Populate(&factory),
	)

	assert.Error(app.Err())
	assert.Nil(factory)
}

func testUnmarshalClaimBuilderError(t *testing.T) {
	var (
		assert  = assert.New(t)
		factory Factory

		app = fx.New(
			fx.Logger(xlog.DiscardPrinter{}),
			fx.Provide(
				config.ProvideViper(
					config.Json(`
						{
							"token": {
								"metadata": [
									{
										"key": "bad"
									}
								],
								"remote": {
									"url": "http://foobar.com"
								}
							}
						}
					`),
				),
				func() key.Registry { return key.NewRegistry(nil) },
				Unmarshal("token"),
			),
			fx.Populate(&factory),
		)
	)

	assert.Error(app.Err())
	assert.Nil(factory)
}

func testUnmarshalFactoryError(t *testing.T) {
	var (
		assert  = assert.New(t)
		factory Factory

		app = fx.New(
			fx.Logger(xlog.DiscardPrinter{}),
			fx.Provide(
				config.ProvideViper(
					config.Json(`
						{
							"token": {
								"alg": "this is not a signing method"
							}
						}
					`),
				),
				func() key.Registry { return key.NewRegistry(nil) },
				Unmarshal("token"),
			),
			fx.Populate(&factory),
		)
	)

	assert.Error(app.Err())
	assert.Nil(factory)
}

func testUnmarshalRequestBuilderError(t *testing.T) {
	var (
		assert  = assert.New(t)
		factory Factory

		app = fx.New(
			fx.Logger(xlog.DiscardPrinter{}),
			fx.Provide(
				config.ProvideViper(
					config.Json(`
						{
							"token": {
								"claims": [
									{
										"key": "bad",
										"header": "X-Bad",
										"parameter": "bad",
										"variable": "bad"
									}
								]
							}
						}
					`),
				),
				func() key.Registry { return key.NewRegistry(nil) },
				Unmarshal("token"),
			),
			fx.Populate(&factory),
		)
	)

	assert.Error(app.Err())
	assert.Nil(factory)
}

func testUnmarshalSuccess(t *testing.T) {
	var (
		assert  = assert.New(t)
		factory Factory

		app = fxtest.New(t,
			fx.Provide(
				config.ProvideViper(
					config.Json(`
						{
							"token": {
								"claims": [
									{
										"key": "static",
										"value": "foo"
									}
								]
							}
						}
					`),
				),
				func() key.Registry { return key.NewRegistry(nil) },
				Unmarshal("token"),
			),
			fx.Populate(&factory),
		)
	)

	assert.NoError(app.Err())
	assert.NotNil(factory)
}

func TestUnmarshal(t *testing.T) {
	t.Run("Error", testUnmarshalError)
	t.Run("ClaimBuilderError", testUnmarshalClaimBuilderError)
	t.Run("FactoryError", testUnmarshalFactoryError)
	t.Run("RequestBuilderError", testUnmarshalRequestBuilderError)
	t.Run("Success", testUnmarshalSuccess)
}
