// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package random

import (
	"bytes"
	"encoding/base64"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testNewBase64NoncerReadError(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)
		empty   bytes.Buffer
		noncer  = NewBase64Noncer(&empty, 0, nil)
	)

	require.NotNil(noncer)
	n, err := noncer.Nonce()
	assert.Empty(n)
	assert.Error(err)
}

func testNewBase64NoncerDefaults(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)
		noncer  = NewBase64Noncer(nil, 0, nil)
	)

	require.NotNil(noncer)
	n, err := noncer.Nonce()
	require.NotEmpty(n)
	require.NoError(err)

	d, err := base64.RawURLEncoding.DecodeString(n)
	require.NoError(err)
	assert.Len(d, 16)
}

func testNewBase64Noncer(t *testing.T, r []byte, encoding *base64.Encoding) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		random = bytes.NewBuffer(r)
		noncer = NewBase64Noncer(random, len(r), encoding)
	)

	require.NotNil(noncer)
	n, err := noncer.Nonce()
	require.NotEmpty(n)
	require.NoError(err)

	d, err := encoding.DecodeString(n)
	require.NoError(err)
	assert.Equal(r, d)
}

func TestNewBase64Noncer(t *testing.T) {
	t.Run("ReadError", testNewBase64NoncerReadError)
	t.Run("Defaults", testNewBase64NoncerDefaults)

	testData := []struct {
		random   []byte
		encoding *base64.Encoding
	}{
		{[]byte{34, 78, 123, 3, 44, 10, 23, 1}, base64.URLEncoding},
		{[]byte{15, 190, 178, 54, 234, 254}, base64.StdEncoding},
	}

	for i, record := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			testNewBase64Noncer(t, record.random, record.encoding)
		})
	}
}
