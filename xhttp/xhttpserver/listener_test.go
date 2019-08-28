package xhttpserver

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testNewListenerTlsError(t *testing.T) {
	var (
		assert = assert.New(t)

		l, err = NewListener(
			context.Background(),
			Options{
				Tls: &Tls{},
			},
			net.ListenConfig{},
		)
	)

	assert.Error(err)
	assert.Nil(l)
}

func testNewListenerListenError(t *testing.T) {
	var (
		assert = assert.New(t)

		l, err = NewListener(
			context.Background(),
			Options{
				Network: "this is not a valid network name",
			},
			net.ListenConfig{},
		)
	)

	assert.Error(err)
	assert.Nil(l)
}

func testNewListenerSimple(t *testing.T) {
	testData := []Options{
		Options{
			DisableTCPKeepAlives: true,
		},
		Options{
			Network:              "tcp4",
			DisableTCPKeepAlives: true,
		},
		Options{
			Network:              "tcp6",
			DisableTCPKeepAlives: true,
		},
	}

	for i, record := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var (
				assert  = assert.New(t)
				require = require.New(t)

				l, err = NewListener(context.Background(), record, net.ListenConfig{})
			)

			require.NoError(err)
			require.NotNil(l)
			defer l.Close()

			assert.IsType((*net.TCPListener)(nil), l)
		})
	}
}

func testNewListenerKeepAlive(t *testing.T) {
	testData := []struct {
		options        Options
		expectedPeriod time.Duration
	}{
		{
			expectedPeriod: defaultTCPKeepAlivePeriod,
		},
		{
			options: Options{
				TCPKeepAlivePeriod: 12 * time.Minute,
			},
			expectedPeriod: 12 * time.Minute,
		},
	}

	for i, record := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var (
				assert  = assert.New(t)
				require = require.New(t)

				l, err = NewListener(context.Background(), record.options, net.ListenConfig{})
			)

			require.NoError(err)
			require.NotNil(l)
			defer l.Close()

			require.IsType(tcpKeepAliveListener{}, l)
			assert.Equal(record.expectedPeriod, l.(tcpKeepAliveListener).period)
		})
	}
}

func testNewListenerTls(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		l, err = NewListener(
			context.Background(),
			Options{
				Tls: &Tls{
					CertificateFile: "server.cert",
					KeyFile:         "server.key",
				},
			},
			net.ListenConfig{},
		)
	)

	require.NoError(err)
	require.NotNil(l)
	defer l.Close()

	// the internal TLS listener type isn't exported, so just make
	// sure we didn't create a listener or decorator of a known type

	_, ok := l.(*net.TCPListener)
	assert.False(ok)

	_, ok = l.(tcpKeepAliveListener)
	assert.False(ok)
}

func TestNewListener(t *testing.T) {
	t.Run("TlsError", testNewListenerTlsError)
	t.Run("ListenError", testNewListenerListenError)
	t.Run("Simple", testNewListenerSimple)
	t.Run("KeepAlive", testNewListenerKeepAlive)
	t.Run("Tls", testNewListenerTls)
}
