package xhttpclient

import (
	"crypto/tls"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTlsConfig(t *testing.T) {
	testData := []struct {
		tls      *Tls
		expected *tls.Config
	}{
		{},
		{
			tls:      &Tls{InsecureSkipVerify: false},
			expected: &tls.Config{InsecureSkipVerify: false},
		},
		{
			tls:      &Tls{InsecureSkipVerify: true},
			expected: &tls.Config{InsecureSkipVerify: true},
		},
	}

	for i, record := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(record.expected, NewTlsConfig(record.tls))
		})
	}
}

func testNewHttpTransportNil(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		rt = NewHttpTransport(nil)
	)

	require.NotNil(rt)
	assert.Equal(new(http.Transport), rt)
}

func testNewHttpTransportDefault(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		rt = NewHttpTransport(new(Transport))
	)

	require.NotNil(rt)
	assert.Equal(new(http.Transport), rt)
}

func testNewHttpTransportFull(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		rt = NewHttpTransport(
			&Transport{
				DisableKeepAlives:      true,
				DisableCompression:     true,
				MaxIdleConns:           245,
				MaxIdleConnsPerHost:    83,
				MaxResponseHeaderBytes: 28765,
				IdleConnTimeout:        11 * time.Second,
				ExpectContinueTimeout:  17 * time.Millisecond,
				TlsHandshakeTimeout:    198 * time.Hour,
				Tls: &Tls{
					InsecureSkipVerify: true,
				},
			},
		)
	)

	require.NotNil(rt)
	require.IsType((*http.Transport)(nil), rt)

	assert.Equal(
		&http.Transport{
			DisableKeepAlives:      true,
			DisableCompression:     true,
			MaxIdleConns:           245,
			MaxIdleConnsPerHost:    83,
			MaxResponseHeaderBytes: 28765,
			IdleConnTimeout:        11 * time.Second,
			ExpectContinueTimeout:  17 * time.Millisecond,
			TLSHandshakeTimeout:    198 * time.Hour,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		rt,
	)
}

func TestNewHttpTransport(t *testing.T) {
	t.Run("Nil", testNewHttpTransportNil)
	t.Run("Default", testNewHttpTransportDefault)
	t.Run("Full", testNewHttpTransportFull)
}

func TestNew(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		c = New(Options{
			Transport: &Transport{
				DisableKeepAlives: true,
			},
			Timeout: 12 * time.Second,
		})
	)

	require.NotNil(c)
	require.IsType((*http.Client)(nil), c)
	assert.Equal(12*time.Second, c.(*http.Client).Timeout)
}

func TestNewCustom(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		rt = new(http.Transport)
		c  = NewCustom(Options{
			Transport: &Transport{
				DisableKeepAlives: true,
			},
			Timeout: 24 * time.Minute,
		}, rt)
	)

	require.NotNil(c)
	require.IsType((*http.Client)(nil), c)
	assert.Equal(24*time.Minute, c.(*http.Client).Timeout)
	assert.Equal(rt, c.(*http.Client).Transport)
}
