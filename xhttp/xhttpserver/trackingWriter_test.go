// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package xhttpserver

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testTrackingWriterBasic(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		next = new(mockResponseWriter)
		tr   = NewTrackingWriter(next)

		actualHeader   = make(http.Header)
		expectedHeader = http.Header{"X-Test": []string{"value1", "value2"}}

		firstWrite  = []byte("test")
		secondWrite = []byte("test again")

		expectedWriteErr = errors.New("expected write error")
	)

	require.NotNil(tr)
	assert.Zero(tr.BytesWritten())
	assert.False(tr.Hijacked())
	assert.Equal(http.StatusOK, tr.StatusCode())

	next.ExpectHeader().Twice().Return(actualHeader)
	next.ExpectWrite(firstWrite).Once().Return(len(firstWrite), error(nil))
	next.ExpectWrite(secondWrite).Once().Return(0, expectedWriteErr)
	next.ExpectWriteHeader(299).Once()

	tr.Header().Add("X-Test", "value1")
	tr.Header().Add("X-Test", "value2")

	c, err := tr.Write(firstWrite)
	assert.Equal(len(firstWrite), c)
	assert.NoError(err)

	c, err = tr.Write(secondWrite)
	assert.Zero(c)
	assert.Equal(expectedWriteErr, err)

	tr.WriteHeader(299)

	assert.Equal(299, tr.StatusCode())
	assert.Equal(expectedHeader, actualHeader)

	assert.Equal(len(firstWrite), tr.BytesWritten())
	assert.False(tr.Hijacked())
	next.AssertExpectations(t)
}

func testTrackingWriterHijack(t *testing.T) {
	t.Run("ImplementsHijacker", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)

			next     = new(mockResponseWriter)
			hijacker = new(mockHijacker)
			tr       = NewTrackingWriter(hijackerWriter{
				ResponseWriter: next,
				Hijacker:       hijacker,
			})

			expectedConn = new(net.IPConn)
			expectedRW   = new(bufio.ReadWriter)
		)

		require.NotNil(tr)
		assert.False(tr.Hijacked())
		hijacker.ExpectHijack().Once().Return(expectedConn, expectedRW, error(nil))

		actualConn, actualRW, err := tr.Hijack()
		assert.Equal(expectedConn, actualConn)
		assert.Equal(expectedRW, actualRW)
		assert.NoError(err)

		assert.True(tr.Hijacked())
		next.AssertExpectations(t)
		hijacker.AssertExpectations(t)
	})

	t.Run("DoesNotImplementHijacker", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)

			next = new(mockResponseWriter)
			tr   = NewTrackingWriter(next)
		)

		require.NotNil(tr)
		assert.False(tr.Hijacked())

		c, rw, err := tr.Hijack()
		assert.Nil(c)
		assert.Nil(rw)
		assert.Equal(ErrHijackerNotSupported, err)

		assert.False(tr.Hijacked())
		next.AssertExpectations(t)
	})
}

func testTrackingWriterPush(t *testing.T) {
	t.Run("ImplementsPusher", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)

			next   = new(mockResponseWriter)
			pusher = new(mockPusher)
			tr     = NewTrackingWriter(pusherWriter{
				ResponseWriter: next,
				Pusher:         pusher,
			})

			expectedTarget      = "test"
			expectedPushOptions = &http.PushOptions{Method: "POST"}
			expectedErr         = errors.New("expected push error")
		)

		require.NotNil(tr)
		pusher.ExpectPush(expectedTarget, expectedPushOptions).Once().Return(expectedErr)

		actualErr := tr.Push(expectedTarget, expectedPushOptions)
		assert.Equal(expectedErr, actualErr)

		next.AssertExpectations(t)
		pusher.AssertExpectations(t)
	})

	t.Run("DoesNotImplementPusher", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)

			next = new(mockResponseWriter)
			tr   = NewTrackingWriter(next)
		)

		require.NotNil(tr)

		err := tr.Push("test", new(http.PushOptions))
		assert.Equal(http.ErrNotSupported, err)

		next.AssertExpectations(t)
	})
}

func testTrackingWriterFlush(t *testing.T) {
	t.Run("ImplementsFlusher", func(t *testing.T) {
		var (
			require = require.New(t)

			next    = new(mockResponseWriter)
			flusher = new(mockFlusher)
			tr      = NewTrackingWriter(flusherWriter{
				ResponseWriter: next,
				Flusher:        flusher,
			})
		)

		require.NotNil(tr)
		flusher.ExpectFlush().Once()

		tr.Flush()

		next.AssertExpectations(t)
		flusher.AssertExpectations(t)
	})

	t.Run("DoesNotImplementFlusher", func(t *testing.T) {
		var (
			require = require.New(t)

			next = new(mockResponseWriter)
			tr   = NewTrackingWriter(next)
		)

		require.NotNil(tr)
		tr.Flush()
		next.AssertExpectations(t)
	})
}

func TestTrackingWriter(t *testing.T) {
	t.Run("Basic", testTrackingWriterBasic)
	t.Run("Hijack", testTrackingWriterHijack)
	t.Run("Push", testTrackingWriterPush)
	t.Run("Flush", testTrackingWriterFlush)
}

func TestNewTrackingWriter(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)
		next    = new(mockResponseWriter)
	)

	tr := NewTrackingWriter(next)
	require.NotNil(tr)

	next.ExpectHeader().Once().Return(http.Header{"X-Test": []string{"value"}})
	assert.Equal(
		http.Header{"X-Test": []string{"value"}},
		tr.Header(),
	)

	assert.Equal(tr, NewTrackingWriter(tr))
	next.AssertExpectations(t)
}

func TestUseTrackingWriter(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		next = http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			assert.Implements((*TrackingWriter)(nil), response)
			response.WriteHeader(299)
		})

		handler  = UseTrackingWriter(next)
		response = httptest.NewRecorder()
	)

	require.NotNil(handler)
	handler.ServeHTTP(response, httptest.NewRequest("GET", "/", nil))
	assert.Equal(299, response.Code)
}
