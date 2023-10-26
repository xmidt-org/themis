// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package xhttpserver

import (
	"net/http"
	"sync/atomic"
)

// busyHandler is the internal http.Handler implementation that wraps another http.Handler
// in concurrent request protection
type busyHandler struct {
	next   http.Handler
	onBusy http.Handler

	maxConcurrentRequests int32
	inFlight              int32
}

func (bh *busyHandler) tryStart() bool {
	if atomic.AddInt32(&bh.inFlight, 1) > bh.maxConcurrentRequests {
		atomic.AddInt32(&bh.inFlight, -1)
		return false
	}

	return true
}

func (bh *busyHandler) end() {
	atomic.AddInt32(&bh.inFlight, -1)
}

func (bh *busyHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if !bh.tryStart() {
		bh.onBusy.ServeHTTP(response, request)
		return
	}

	defer bh.end()
	bh.next.ServeHTTP(response, request)
}

// Busy is an Alice-style decorator that enforces a maximum number of concurrent HTTP transactions
type Busy struct {
	MaxConcurrentRequests int
	OnBusy                http.Handler
}

func (b Busy) Then(next http.Handler) http.Handler {
	if b.MaxConcurrentRequests < 1 {
		return next
	}

	bh := &busyHandler{
		maxConcurrentRequests: int32(b.MaxConcurrentRequests),
		next:                  next,
	}

	if b.OnBusy != nil {
		bh.onBusy = b.OnBusy
	} else {
		bh.onBusy = Constant{StatusCode: http.StatusTooManyRequests}.NewHandler()
	}

	return bh
}

func (b Busy) ThenFunc(next http.HandlerFunc) http.Handler {
	return b.Then(next)
}
