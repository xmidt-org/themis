// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package token

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/xmidt-org/themis/key"
	"github.com/xmidt-org/themis/random"
	"github.com/xmidt-org/themis/random/randomtest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testNewFactoryInvalidAlg(t *testing.T) {
	var (
		assert = assert.New(t)
		f, err = NewFactory(
			Options{
				Alg: "this is not a signing method",
			},
			ClaimBuilders{},
			key.NewRegistry(nil),
		)
	)

	assert.Nil(f)
	assert.Error(err)
}

func testNewFactoryInvalidKeyType(t *testing.T) {
	var (
		assert = assert.New(t)
		f, err = NewFactory(
			Options{
				Key: key.Descriptor{
					Type: "this is not a valid key type",
				},
			},
			ClaimBuilders{},
			key.NewRegistry(nil),
		)
	)

	assert.Nil(f)
	assert.Error(err)
}

func testNewFactoryWithoutReponseClaimsSuccess(t *testing.T) {
	var (
		assert   = assert.New(t)
		require  = require.New(t)
		registry = key.NewRegistry(rand.Reader)
		noncer   = random.NewBase64Noncer(rand.Reader, 128, nil)

		cb = ClaimBuilders{
			nonceClaimBuilder{n: noncer},
		}
	)

	factory, err := NewFactory(Options{
		Alg: "RS256",
		Key: key.Descriptor{
			Kid:  "test",
			Bits: 512,
		},
		Nonce: true,
	}, cb, registry)

	require.NoError(err)
	require.NotNil(factory)

	token, err := factory.NewToken(context.Background(), new(Request), nil)
	require.NoError(err)
	assert.True(len(token) > 0)
}

func testNewFactoryWithReponseClaimsSuccess(t *testing.T) {
	var (
		assert         = assert.New(t)
		require        = require.New(t)
		registry       = key.NewRegistry(rand.Reader)
		noncer         = new(randomtest.Noncer)
		expectedClaims = map[string]interface{}{"jti": "deadbeef"}
		cb             = ClaimBuilders{
			nonceClaimBuilder{n: noncer},
		}
	)

	noncer.ExpectNonce().Return("deadbeef", error(nil)).Once()
	factory, err := NewFactory(Options{
		Alg: "RS256",
		Key: key.Descriptor{
			Kid:  "test",
			Bits: 512,
		},
		Nonce: true,
	}, cb, registry)

	require.NoError(err)
	require.NotNil(factory)

	resp := Response{Claims: make(map[string]interface{})}
	token, err := factory.NewToken(context.Background(), new(Request), resp.Claims)
	require.NoError(err)
	assert.True(len(token) > 0)
	assert.Equal(expectedClaims, resp.Claims)
}

func TestNewFactory(t *testing.T) {
	t.Run("InvalidAlg", testNewFactoryInvalidAlg)
	t.Run("InvalidKeyType", testNewFactoryInvalidKeyType)
	t.Run("WithoutReponseClaims Success", testNewFactoryWithoutReponseClaimsSuccess)
	t.Run("WithReponseClaims Success", testNewFactoryWithReponseClaimsSuccess)
}
