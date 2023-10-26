// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package xhttpserver

import (
	"bytes"
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/xmidt-org/sallust"
)

func testOnStartNewListenerError(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		output bytes.Buffer

		s       = new(mockServer)
		onStart = OnStart(
			Options{
				Tls: &Tls{},
			},
			s,
			sallust.Default(),
			func() {
				assert.Fail("onExit should not have been called")
			},
		)
	)

	require.NotNil(onStart)
	err := onStart(context.Background())
	assert.Error(err)
	assert.Zero(output.Len())
	s.AssertExpectations(t)
}

func testOnStartSuccess(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		onExitCalled = make(chan struct{})
		serve        = make(chan net.Listener, 1)
		s            = new(mockServer)
		onStart      = OnStart(
			Options{},
			s,
			sallust.Default(),
			func() {
				close(onExitCalled)
			},
		)
	)

	require.NotNil(onStart)
	s.ExpectServe(mock.MatchedBy(func(net.Listener) bool { return true })).Once().Return(http.ErrServerClosed).
		Run(func(arguments mock.Arguments) {
			serve <- arguments.Get(0).(net.Listener)
		})

	assert.NoError(onStart(context.Background()))
	select {
	case l := <-serve:
		l.Close() // passing, but clean up after ourselves
	case <-time.After(time.Second):
		assert.Fail("Serve was not called")
	}

	select {
	case <-onExitCalled:
		// passing
	case <-time.After(time.Second):
		assert.Fail("onExit was not called")
	}

	s.AssertExpectations(t)
}

func TestOnStart(t *testing.T) {
	t.Run("NewListenerError", testOnStartNewListenerError)
	t.Run("Success", testOnStartSuccess)
}

func TestOnStop(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		expectedErr = errors.New("expected shutdown error")
		s           = new(mockServer)
		onStop      = OnStop(s, sallust.Default())
	)

	require.NotNil(onStop)
	s.ExpectShutdown(mock.MatchedBy(func(context.Context) bool { return true })).Once().Return(expectedErr)
	assert.Equal(expectedErr, onStop(context.Background()))

	s.AssertExpectations(t)
}
