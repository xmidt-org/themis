package key

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateRSAPair(t *testing.T) {
	testBits := []int{
		0,
		128,
		256,
		512,
		1024,
	}

	for _, bits := range testBits {
		t.Run(fmt.Sprintf("bits=%d", bits), func(t *testing.T) {
			var (
				assert  = assert.New(t)
				require = require.New(t)

				pair, err = GenerateRSAPair("test", rand.Reader, bits)
			)

			require.NoError(err)
			require.NotNil(pair)

			assert.Equal("test", pair.KID())

			key, ok := pair.Sign().(*rsa.PrivateKey)
			require.True(ok)
			require.NotNil(key)

			var output bytes.Buffer
			c, err := pair.WriteVerifyPEMTo(&output)
			require.NoError(err)
			assert.True(c > 0)
		})
	}
}

func testGenerateECDSAPairSuccess(t *testing.T, bits int) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		pair, err = GenerateECDSAPair("test", rand.Reader, bits)
	)

	require.NoError(err)
	require.NotNil(pair)

	assert.Equal("test", pair.KID())

	key, ok := pair.Sign().(*ecdsa.PrivateKey)
	require.True(ok)
	require.NotNil(key)

	var output bytes.Buffer
	c, err := pair.WriteVerifyPEMTo(&output)
	require.NoError(err)
	assert.True(c > 0)
}

func TestGenerateECDSAPair(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		goodBits := []int{
			0,
			224,
			256,
			384,
			512,
		}

		for _, bits := range goodBits {
			t.Run(fmt.Sprintf("bits=%d", bits), func(t *testing.T) {
				testGenerateECDSAPairSuccess(t, bits)
			})
		}
	})

	t.Run("InvalidBits", func(t *testing.T) {
		var (
			assert    = assert.New(t)
			pair, err = GenerateECDSAPair("test", rand.Reader, 111)
		)

		assert.Nil(pair)
		assert.Error(err)
	})
}

func TestGenerateSecretPair(t *testing.T) {
	testBits := []int{
		0,
		128,
		256,
		512,
		1024,
	}

	for _, bits := range testBits {
		t.Run(fmt.Sprintf("bits=%d", bits), func(t *testing.T) {
			var (
				assert    = assert.New(t)
				require   = require.New(t)
				pair, err = GenerateSecretPair("test", rand.Reader, bits)
			)

			require.NoError(err)
			require.NotNil(pair)
			assert.Equal("test", pair.KID())

			key, ok := pair.Sign().([]byte)
			require.True(ok)
			require.NotNil(key)

			if bits <= 0 {
				assert.Len(key, DefaultSecretBits)
			} else {
				assert.Len(key, bits)
			}

			var output bytes.Buffer
			c, err := pair.WriteVerifyPEMTo(&output)
			require.NoError(err)
			assert.True(c > 0)
		})
	}
}
