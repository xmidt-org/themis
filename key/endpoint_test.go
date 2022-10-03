package key

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEndpoint(t *testing.T) {
	t.Run("KeyFound", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)

			registry = NewRegistry(nil)
			endpoint = NewEndpoint(registry)
		)

		require.NotNil(endpoint)
		_, err := registry.Register(Descriptor{Kid: "test"})
		require.NoError(err)

		result, err := endpoint(context.Background(), "test")
		require.NoError(err)
		require.NotNil(result)

		pair, ok := result.(Pair)
		require.True(ok)
		require.NotNil(pair)

		assert.Equal("test", pair.KID())
	})

	t.Run("KeyNotFound", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)

			registry = NewRegistry(nil)
			endpoint = NewEndpoint(registry)
		)

		require.NotNil(endpoint)

		result, err := endpoint(context.Background(), "test")
		require.Error(err)
		require.Nil(result)
		var keyNotFoundError KeyNotFoundError
		assert.ErrorAs(err, &keyNotFoundError)
		assert.Equal(http.StatusNotFound, keyNotFoundError.StatusCode())
	})
}
