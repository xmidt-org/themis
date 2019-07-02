package key

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRegistryRegister(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		registry = NewRegistry(rand.Reader)
	)

	require.NotNil(registry)

	pair, err := registry.Register(Descriptor{
		Kid:  "test",
		Bits: 512,
	})

	require.NoError(err)
	require.NotNil(pair)
	assert.Equal("test", pair.KID())

	key, ok := pair.Sign().(*rsa.PrivateKey)
	require.True(ok)
	require.NotNil(key)

	existing, ok := registry.Get("test")
	require.True(ok)
	require.NotNil(existing)
	assert.Equal(pair, existing)
}

func TestRegistry(t *testing.T) {
	t.Run("Register", testRegistryRegister)
}
