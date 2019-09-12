package xhttpserver

import (
	"bytes"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddressKey(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(addressKey, AddressKey())
}

func TestServerKey(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(serverKey, ServerKey())
}

func TestNewServerLogger(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)

			output bytes.Buffer
			base   = log.NewJSONLogger(&output)
			logger = NewServerLogger("test", Options{}, base)
		)

		require.NotNil(logger)
		logger.Log("foo", "bar")
		assert.Greater(output.Len(), 0)
		assert.Contains(output.String(), "test")
	})

	t.Run("Extra", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)

			output bytes.Buffer
			base   = log.NewJSONLogger(&output)
			logger = NewServerLogger("test", Options{}, base, "port", 1234)
		)

		require.NotNil(logger)
		logger.Log("foo", "bar")
		assert.Greater(output.Len(), 0)
		assert.Contains(output.String(), "test")
		assert.Contains(output.String(), "port")
		assert.Contains(output.String(), "1234")
	})
}
