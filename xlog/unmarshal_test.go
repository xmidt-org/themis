package xlog

import (
	"testing"

	"github.com/xmidt-org/themis/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

func testUnmarshalSuccess(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		logger *zap.Logger

		app = fxtest.New(t,
			fx.Provide(
				config.ProvideViper(
					config.Json(`
						{
							"log": {
								"file": "stdout",
								"level": "DEBUG"
							}
						}`,
					),
				),
				Unmarshal("log"),
			),
			fx.Populate(&logger),
		)
	)

	require.NoError(app.Err())
	assert.NotNil(logger)
}

func testUnmarshalWithBufferedPrinter(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		logger  *zap.Logger
		printer *BufferedPrinter

		app = fxtest.New(t,
			Logger(),
			fx.Provide(
				config.ProvideViper(
					config.Json(`
						{
							"log": {
								"file": "stdout",
								"level": "ERROR"
							}
						}`,
					),
				),
				Unmarshal("log"),
			),
			fx.Populate(&logger, &printer),
		)
	)

	require.NoError(app.Err())
	assert.NotNil(logger)

	require.NotNil(printer)
	assert.NotNil(printer.logger)
}

func testUnmarshalFailure(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		logger *zap.Logger

		app = fx.New(
			fx.Logger(DiscardPrinter{}),
			fx.Provide(
				config.ProvideViper(
					config.Json(`
						{
							"log": {
								"file": "stdout",
								"maxBackups": "this is not a valid int"
							}
						}`,
					),
				),
				Unmarshal("log"),
			),
			fx.Populate(&logger),
		)
	)

	require.Error(app.Err())
	assert.Nil(logger)
}

func TestUnmarshal(t *testing.T) {
	t.Run("Success", testUnmarshalSuccess)
	t.Run("WithBufferedPrinter", testUnmarshalWithBufferedPrinter)
	t.Run("Failure", testUnmarshalFailure)
}
