package xmetricshttp

import (
	"net/http"
	"time"
)

// InstrumentHandlerCounter provides a simple count of HTTP requests, usually labelled in some interesting way
// by the Reporter (e.g. code and method).
type InstrumentHandlerCounter struct {
	// Reporter is e counter-based reporter that receives postive values.  If not set, no metric is gathered.
	Reporter Reporter
}

func (ihc InstrumentHandlerCounter) Then(next http.Handler) http.Handler {
	if ihc.Reporter == nil {
		return next
	}

	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		next.ServeHTTP(response, request)
		ihc.Reporter.Report(response, request, 1.0)
	})
}

// InstrumentHandlerDuration provides request duration metrics
type InstrumentHandlerDuration struct {
	// Reporter is an observer-based reporter that receives duration values.  If not set, no metric is gathered.
	Reporter Reporter

	// Now is the optional strategy for obtaining the system time.  If not supplied, time.Now is used.
	Now func() time.Time
}

func (ihd InstrumentHandlerDuration) Then(next http.Handler) http.Handler {
	if ihd.Reporter == nil {
		return next
	}

	now := ihd.Now
	if now == nil {
		now = time.Now
	}

	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		start := now()
		next.ServeHTTP(response, request)
		ihd.Reporter.Report(
			response,
			request,
			float64(now().Sub(start)),
		)
	})
}

// InstrumentHandlerInFlight records how many current HTTP transactions are being executed by an http.Handler
type InstrumentHandlerInFlight struct {
	// Reporter is a gauge-based reporter that receives postive and negative deltas.  If not set, no metric is gathered.
	Reporter Reporter
}

func (ihif InstrumentHandlerInFlight) Then(next http.Handler) http.Handler {
	if ihif.Reporter == nil {
		return next
	}

	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		defer ihif.Reporter.Report(response, request, -1.0)
		ihif.Reporter.Report(response, request, 1.0)
		next.ServeHTTP(response, request)
	})
}
