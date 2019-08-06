package token

import (
	"context"
	"crypto/rand"
	"key"
	"random"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testNewFactoryInvalidAlg(t *testing.T) {
	var (
		assert = assert.New(t)
		f, err = NewFactory(
			ClaimBuilders{},
			key.NewRegistry(nil),
			Options{
				Alg: "this is not a signing method",
			},
		)
	)

	assert.Nil(f)
	assert.Error(err)
}

func testNewFactoryInvalidKeyType(t *testing.T) {
	var (
		assert = assert.New(t)
		f, err = NewFactory(
			ClaimBuilders{},
			key.NewRegistry(nil),
			Options{
				Key: key.Descriptor{
					Type: "this is not a valid key type",
				},
			},
		)
	)

	assert.Nil(f)
	assert.Error(err)
}

func testNewFactorySuccess(t *testing.T) {
	var (
		assert   = assert.New(t)
		require  = require.New(t)
		registry = key.NewRegistry(rand.Reader)
		noncer   = random.NewBase64Noncer(rand.Reader, 128, nil)

		cb = ClaimBuilders{
			nonceClaimBuilder{n: noncer},
		}
	)

	factory, err := NewFactory(cb, registry, Options{
		Alg: "RS256",
		Key: key.Descriptor{
			Kid:  "test",
			Bits: 512,
		},
		Nonce: true,
	})

	require.NoError(err)
	require.NotNil(factory)

	token, err := factory.NewToken(context.Background(), new(Request))
	require.NoError(err)
	assert.True(len(token) > 0)
}

func TestNewFactory(t *testing.T) {
	t.Run("InvalidAlg", testNewFactoryInvalidAlg)
	t.Run("InvalidKeyType", testNewFactoryInvalidKeyType)
	t.Run("Success", testNewFactorySuccess)
}
