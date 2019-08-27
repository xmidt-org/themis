package token

import (
	"errors"
	"testing"

	"github.com/xmidt-org/themis/config"
	"github.com/xmidt-org/themis/config/configtest"
	"github.com/xmidt-org/themis/key"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func testUnmarshalError(t *testing.T) {
	var (
		assert = assert.New(t)

		unmarshaller = new(configtest.Unmarshaller)
		factory      Factory
	)

	unmarshaller.ExpectUnmarshalKey("token", mock.AnythingOfType("*token.Options")).Once().Return(errors.New("expected"))

	app := fx.New(
		fx.Logger(config.DiscardPrinter{}),
		fx.Provide(
			func() config.Unmarshaller { return unmarshaller },
			func() key.Registry { return key.NewRegistry(nil) },
			Unmarshal("token"),
		),
		fx.Populate(&factory),
	)

	assert.Error(app.Err())
	assert.Nil(factory)
	unmarshaller.AssertExpectations(t)
}

func testUnmarshalClaimBuilderError(t *testing.T) {
	var (
		assert = assert.New(t)

		v = configtest.LoadJson(t,
			`
				{
					"token": {
						"metadata": {
							"bad": {
							}
						},
						"remote": {
							"url": "http://foobar.com"
						}
					}
				}
			`,
		)

		factory Factory

		app = fx.New(
			fx.Logger(config.DiscardPrinter{}),
			fx.Provide(
				func() config.Unmarshaller { return config.ViperUnmarshaller{Viper: v} },
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
		assert = assert.New(t)

		v = configtest.LoadJson(t,
			`
				{
					"token": {
						"alg": "this is not a signing method"
					}
				}
			`,
		)

		factory Factory

		app = fx.New(
			fx.Logger(config.DiscardPrinter{}),
			fx.Provide(
				func() config.Unmarshaller { return config.ViperUnmarshaller{Viper: v} },
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
		assert = assert.New(t)

		v = configtest.LoadJson(t,
			`
				{
					"token": {
						"claims": {
							"bad": {
								"header": "X-Bad",
								"parameter": "bad",
								"variable": "bad"
							}
						}
					}
				}
			`,
		)

		factory Factory

		app = fx.New(
			fx.Logger(config.DiscardPrinter{}),
			fx.Provide(
				func() config.Unmarshaller { return config.ViperUnmarshaller{Viper: v} },
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
		assert = assert.New(t)

		v = configtest.LoadJson(t,
			`
				{
					"token": {
						"claims": {
							"static": {
								"value": "foo"
							}
						}
					}
				}
			`,
		)

		factory Factory

		app = fxtest.New(t,
			fx.Provide(
				func() config.Unmarshaller { return config.ViperUnmarshaller{Viper: v} },
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
