package xlog

import (
	"bytes"
	"config"
	"config/configtest"
	"errors"
	"strings"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func testUnmarshalSuccess(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		v      = viper.New()
		logger log.Logger
	)

	v.SetConfigType("json")
	require.NoError(
		v.ReadConfig(strings.NewReader(`
		{
			"log": {
				"file": "stdout",
				"level": "DEBUG"
			}
		}
	`)),
	)

	app := fxtest.New(t,
		fx.Provide(
			func() config.Unmarshaller {
				return config.ViperUnmarshaller{Viper: v}
			},
			Unmarshal("log"),
		),
		fx.Populate(&logger),
	)

	require.NoError(app.Err())
	assert.NotNil(logger)
}

func testUnmarshalFailure(t *testing.T) {
	var (
		assert       = assert.New(t)
		unmarshaller = new(configtest.Unmarshaller)

		logger log.Logger
	)

	unmarshaller.ExpectUnmarshalKey("log", mock.AnythingOfType("*xlog.Options")).Once().Return(errors.New("expected"))
	app := fx.New(
		fx.Logger(config.DiscardPrinter{}),
		fx.Provide(
			func() config.Unmarshaller {
				return unmarshaller
			},
			Unmarshal("log"),
		),
		fx.Populate(&logger),
	)

	assert.Error(app.Err())
	assert.Nil(logger)
	unmarshaller.AssertExpectations(t)
}

func TestUnmarshal(t *testing.T) {
	t.Run("Success", testUnmarshalSuccess)
	t.Run("Failure", testUnmarshalFailure)
}

func testUnmarshallerWithPrinter(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		expected  bytes.Buffer
		component log.Logger

		v = viper.New()
	)

	v.SetConfigType("json")
	require.NoError(
		v.ReadConfig(strings.NewReader(`
		{
			"log": {
				"file": "stdout",
				"level": "DEBUG"
			}
		}
	`)),
	)

	optioner := Unmarshaller("log", func(_ log.Logger, e config.Environment) fx.Printer {
		return Printer{Logger: log.NewJSONLogger(&expected)}
	})

	require.NotNil(optioner)

	app := fxtest.New(t,
		optioner(config.Environment{Unmarshaller: config.ViperUnmarshaller{Viper: v}}),
		fx.Populate(&component),
	)

	require.NoError(app.Err())
	assert.NotNil(component)
	assert.Greater(expected.Len(), 0)
}

func testUnmarshallerNoPrinter(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		component log.Logger

		v = viper.New()
	)

	v.SetConfigType("json")
	require.NoError(
		v.ReadConfig(strings.NewReader(`
		{
			"log": {
				"file": "stdout",
				"level": "DEBUG"
			}
		}
	`)),
	)

	optioner := Unmarshaller("log", nil)
	require.NotNil(optioner)

	app := fxtest.New(t,
		optioner(config.Environment{Unmarshaller: config.ViperUnmarshaller{Viper: v}}),
		fx.Populate(&component),
	)

	require.NoError(app.Err())
	assert.NotNil(component)
}

func testUnmarshallerFailure(t *testing.T) {
	var (
		assert       = assert.New(t)
		require      = require.New(t)
		unmarshaller = new(configtest.Unmarshaller)

		logger log.Logger
	)

	optioner := Unmarshaller("log", nil)
	require.NotNil(optioner)

	unmarshaller.ExpectUnmarshalKey("log", mock.AnythingOfType("*xlog.Options")).Once().Return(errors.New("expected"))
	app := fx.New(
		fx.Logger(config.DiscardPrinter{}),
		fx.Provide(
			func() config.Unmarshaller {
				return unmarshaller
			},
			optioner(config.Environment{Unmarshaller: unmarshaller}),
		),
		fx.Populate(&logger),
	)

	assert.Error(app.Err())
	assert.Nil(logger)
	unmarshaller.AssertExpectations(t)
}

func TestUnmarshaller(t *testing.T) {
	t.Run("WithPrinter", testUnmarshallerWithPrinter)
	t.Run("NoPrinter", testUnmarshallerWithPrinter)
	t.Run("Failure", testUnmarshallerFailure)
}
