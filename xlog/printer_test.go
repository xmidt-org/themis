package xlog

import (
	"bytes"
	"errors"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestPrinter(t *testing.T) {
	var (
		assert = assert.New(t)
		output bytes.Buffer
		logger = log.NewJSONLogger(&output)

		p = Printer{Logger: logger}
	)

	p.Printf("test: %d", 123)
	assert.Greater(output.Len(), 0)
}

func testBufferedPrinterBasic(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)
		printer *BufferedPrinter
		dummy   string

		output1 bytes.Buffer
		logger1 = log.NewJSONLogger(&output1)

		output2 bytes.Buffer
		logger2 = log.NewJSONLogger(&output2)

		app = fxtest.New(t,
			Logger(),
			fx.Provide(
				func() string {
					return "dummy component"
				},
			),
			fx.Invoke(
				func(bp *BufferedPrinter) {
					assert.Greater(bp.Len(), 0)
					bp.SetLogger(logger1)
					assert.Zero(bp.Len())
					byteCount := output1.Len()
					assert.Greater(byteCount, 0)

					bp.SetLogger(logger2)
					assert.Equal(byteCount, output1.Len())
					assert.Zero(output2.Len())
				},
			),
			fx.Populate(&printer, &dummy),
		)
	)

	require.NoError(app.Err())
	require.NotNil(printer)

	assert.Empty(printer.messages)
	assert.NotNil(printer.logger)

	app.RequireStart()
	assert.Empty(printer.messages)
	assert.NotNil(printer.logger)
}

func testBufferedPrinterHandleError(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)
		dummy   string

		output bytes.Buffer
		logger = log.NewJSONLogger(&output)

		app = fx.New(
			Logger(),
			fx.Provide(
				func() (string, error) {
					return "uh oh!", errors.New("expected")
				},
			),
			fx.Invoke(
				func(bp *BufferedPrinter) {
					bp.SetLogger(logger)
				},
			),
			fx.Populate(&dummy),
		)
	)

	require.Error(app.Err())
	assert.Contains(output.String(), "expected")
}

func TestBufferedPrinter(t *testing.T) {
	t.Run("Basic", testBufferedPrinterBasic)
	t.Run("HandleError", testBufferedPrinterHandleError)
}
