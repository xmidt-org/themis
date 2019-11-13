package xhttpserver

import (
	"context"
	"crypto/tls"
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
	if l != nil {
		l.Close()
	}
}

func testNewListenerSimple(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

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
	}()

	c, err := net.DialTimeout("tcp", l.Addr().String(), 5*time.Second)
	require.NoError(err)
	require.NotNil(c)

	defer c.Close()
	acceptWait.Wait()
}

func testNewListenerSimpleTls(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		tlsConfig = &tls.Config{
			GetCertificate: newGetCertificate(t),
		}

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
	}()

	c, err := net.DialTimeout("tcp", l.Addr().String(), 5*time.Second)
	require.NoError(err)
	require.NotNil(c)

	defer c.Close()
	acceptWait.Wait()
}

func TestNewListener(t *testing.T) {
	t.Run("InvalidAddress", testNewListenerInvalidAddress)
	t.Run("Simple", testNewListenerSimple)
	t.Run("SimpleTls", testNewListenerSimpleTls)
}
