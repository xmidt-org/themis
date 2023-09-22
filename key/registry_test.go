// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package key

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRegistryValidDescriptor(t *testing.T, d Descriptor) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		registry = NewRegistry(nil)
	)

	require.NotNil(registry)

	pair, err := registry.Register(d)
	require.NoError(err)
	require.NotNil(pair)
	assert.Equal(d.Kid, pair.KID())

	pair, ok := registry.Get(d.Kid)
	assert.True(ok)
	assert.NotNil(pair)

	switch d.Type {
	case "":
		fallthrough
	case KeyTypeRSA:
		_, ok := pair.Sign().(*rsa.PrivateKey)
		assert.True(ok)
	case KeyTypeECDSA:
		_, ok := pair.Sign().(*ecdsa.PrivateKey)
		assert.True(ok)
	case KeyTypeSecret:
		_, ok := pair.Sign().([]byte)
		assert.True(ok)
	default:
		assert.Fail("Unknown key type", d.Type)
	}

	// idempotency
	pair, err = registry.Register(d)
	assert.Error(err)
	assert.Nil(pair)
}

func testRegistryInvalidDescriptor(t *testing.T, d Descriptor) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		registry = NewRegistry(nil)
	)

	require.NotNil(registry)

	pair, err := registry.Register(d)
	require.Error(err)
	require.Nil(pair)

	pair, ok := registry.Get(d.Kid)
	assert.False(ok)
	assert.Nil(pair)
}

func TestRegistry(t *testing.T) {
	t.Run("Register", func(t *testing.T) {
		t.Run("ValidDescriptor", func(t *testing.T) {
			testData := []Descriptor{
				Descriptor{Kid: "test"},
				Descriptor{Kid: "test", Type: "rsa"},
				Descriptor{Kid: "test", Type: "rsa", File: "test.pkcs1.pem"},
				Descriptor{Kid: "test", Type: "rsa", File: "test.pkcs8.pem"},
				Descriptor{Kid: "test", Type: "ecdsa"},
				Descriptor{Kid: "test", Type: "secret"},
			}

			for i, d := range testData {
				t.Run(strconv.Itoa(i), func(t *testing.T) {
					testRegistryValidDescriptor(t, d)
				})
			}
		})

		t.Run("InvalidDescriptor", func(t *testing.T) {
			testData := []Descriptor{
				Descriptor{File: "nosuch"},
				Descriptor{Type: "not a valid key type"},
			}

			for i, d := range testData {
				t.Run(strconv.Itoa(i), func(t *testing.T) {
					testRegistryInvalidDescriptor(t, d)
				})
			}
		})
	})
}
