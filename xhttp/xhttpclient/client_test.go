// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package xhttpclient

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xmidt-org/arrange/arrangehttp"
)

func TestNew(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)
	)
	c, err := New(Options{
		HTTPClient: arrangehttp.ClientConfig{
			Transport: arrangehttp.TransportConfig{DisableKeepAlives: true},
			Timeout:   12 * time.Second,
		}})
	require.NoError(err)
	require.NotNil(c)
	require.IsType((*http.Client)(nil), c)
	assert.Equal(12*time.Second, c.(*http.Client).Timeout)
}

func TestNewCustom(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		rt = new(http.Transport)
	)

	c, err := NewCustom(Options{
		HTTPClient: arrangehttp.ClientConfig{
			Transport: arrangehttp.TransportConfig{DisableKeepAlives: true},
			Timeout:   24 * time.Minute,
		}}, rt)
	require.NoError(err)
	require.NotNil(c)
	require.IsType((*http.Client)(nil), c)
	assert.Equal(24*time.Minute, c.(*http.Client).Timeout)
	assert.Equal(rt, c.(*http.Client).Transport)
}
