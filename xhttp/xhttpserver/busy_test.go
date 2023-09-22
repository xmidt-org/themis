// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package xhttpserver

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func testBusyNoDecoration(t *testing.T) {
	var (
		assert = assert.New(t)

		next = Constant{}.NewHandler()
		busy = Busy{}.Then(next)
	)

	assert.Equal(next, busy)
}

func testBusyDefaultOnBusy(t *testing.T) {
	var (
		assert = assert.New(t)

		nextFinish      = new(sync.WaitGroup)
		nextInServeHTTP = make(chan struct{})
		nextBlock       = make(chan struct{})
		next            = func(response http.ResponseWriter, request *http.Request) {
			close(nextInServeHTTP)
			<-nextBlock
			response.WriteHeader(288)
		}

		busy = Busy{MaxConcurrentRequests: 1}.ThenFunc(next)
	)

	nextFinish.Add(1)

	go func() {
		defer nextFinish.Done()
		response := httptest.NewRecorder()
		busy.ServeHTTP(response, httptest.NewRequest("GET", "/", nil))
		assert.Equal(288, response.Code)
	}()

	select {
	case <-nextInServeHTTP:
		// passing
	case <-time.After(time.Second):
		assert.Fail("Busy did not call next.ServeHTTP")
	}

	response := httptest.NewRecorder()
	busy.ServeHTTP(response, httptest.NewRequest("GET", "/", nil))
	assert.Equal(http.StatusTooManyRequests, response.Code)

	close(nextBlock)
	nextFinish.Wait()
}

func testBusyCustomOnBusy(t *testing.T) {
	var (
		assert = assert.New(t)

		nextFinish      = new(sync.WaitGroup)
		nextInServeHTTP = make(chan struct{})
		nextBlock       = make(chan struct{})
		next            = func(response http.ResponseWriter, request *http.Request) {
			close(nextInServeHTTP)
			<-nextBlock
			response.WriteHeader(288)
		}

		busy = Busy{
			MaxConcurrentRequests: 1,
			OnBusy:                Constant{StatusCode: 476}.NewHandler(),
		}.ThenFunc(next)
	)

	nextFinish.Add(1)

	go func() {
		defer nextFinish.Done()
		response := httptest.NewRecorder()
		busy.ServeHTTP(response, httptest.NewRequest("GET", "/", nil))
		assert.Equal(288, response.Code)
	}()

	select {
	case <-nextInServeHTTP:
		// passing
	case <-time.After(time.Second):
		assert.Fail("Busy did not call next.ServeHTTP")
	}

	response := httptest.NewRecorder()
	busy.ServeHTTP(response, httptest.NewRequest("GET", "/", nil))
	assert.Equal(476, response.Code)

	close(nextBlock)
	nextFinish.Wait()
}

func TestBusy(t *testing.T) {
	t.Run("NoDecoration", testBusyNoDecoration)
	t.Run("DefaultOnBusy", testBusyDefaultOnBusy)
	t.Run("DefaultCustomBusy", testBusyCustomOnBusy)
}
