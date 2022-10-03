package xhttpserver

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testNewListenerInvalidAddress(t *testing.T) {
	assert := assert.New(t)
	l, err := NewListener(context.Background(), Options{Address: "invalid address"}, net.ListenConfig{}, nil)
	assert.Error(err)
	if !assert.Nil(l) {
		l.Close()
	}
}

func testNewListenerNonTLS(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		expectedMessage = []byte("hello, world")

		listenCtx, listenCancel = context.WithTimeout(context.Background(), time.Minute)
		acceptWait              sync.WaitGroup
	)

	defer listenCancel()
	l, err := NewListener(listenCtx, Options{Address: ":0"}, net.ListenConfig{}, nil)
	require.NoError(err)
	require.NotNil(l)

	defer l.Close()
	acceptWait.Add(1)

	go func() {
		defer acceptWait.Done()
		c, err := l.Accept()
		if !assert.NoError(err) {
			if c != nil {
				c.Close()
			}

			return
		}

		defer c.Close()
		assert.IsType((*net.TCPConn)(nil), c)
		c.Write(expectedMessage)
	}()

	c, err := net.DialTimeout("tcp", l.Addr().String(), 5*time.Second)
	require.NoError(err)
	require.NotNil(c)

	defer c.Close()
	acceptWait.Wait()

	actualMessage := make([]byte, len(expectedMessage))
	n, err := io.ReadFull(c, actualMessage)
	assert.Equal(len(actualMessage), n)
	assert.NoError(err)
	assert.Equal(expectedMessage, actualMessage)
}

func testNewListenerTLS(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		expectedMessage = []byte("hello, world")
		tlsConfig       = addServerCertificate(t, nil)

		listenCtx, listenCancel = context.WithTimeout(context.Background(), time.Minute)
		acceptWait              sync.WaitGroup
	)

	defer listenCancel()
	l, err := NewListener(listenCtx, Options{Address: ":0"}, net.ListenConfig{}, tlsConfig)
	require.NoError(err)
	require.NotNil(l)

	defer l.Close()
	acceptWait.Add(1)

	go func() {
		defer acceptWait.Done()
		c, err := l.Accept()
		if !assert.NoError(err) {
			if c != nil {
				c.Close()
			}

			return
		}

		defer c.Close()
		assert.IsType((*tls.Conn)(nil), c)
		assert.Implements((*TlsConn)(nil), c)
		c.Write(expectedMessage)
	}()

	c, err := tls.Dial("tcp", l.Addr().String(), &tls.Config{InsecureSkipVerify: true}) // nolint: gosec
	require.NoError(err)
	require.NotNil(c)

	defer c.Close()
	acceptWait.Wait()

	actualMessage := make([]byte, len(expectedMessage))
	n, err := io.ReadFull(c, actualMessage)
	assert.Equal(len(actualMessage), n)
	assert.NoError(err)
	assert.Equal(expectedMessage, actualMessage)
}

func TestNewListener(t *testing.T) {
	t.Run("InvalidAddress", testNewListenerInvalidAddress)
	t.Run("NonTLS", testNewListenerNonTLS)
	t.Run("TLS", testNewListenerTLS)
}
