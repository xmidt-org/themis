package xmetricshttp

import (
	"net/http"
	"time"
	"xmetrics"
)

// InstrumentHandlerCounter provides a simple count metric of HTTP transactions
type InstrumentHandlerCounter struct {
	Reporter AdderReporter
	Labeller ServerLabeller
}

func (ihc InstrumentHandlerCounter) Then(next http.Handler) http.Handler {
	if ihc.Reporter == nil {
		return next
	}

	labeller := ihc.Labeller
	if ihc.Labeller == nil {
		labeller = EmptyLabeller{}
	}

	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		next.ServeHTTP(response, request)
		var l xmetrics.Labels
		labeller.ServerLabels(response, request, &l)
		ihc.Reporter.Report(&l, 1.0)
	})
}

// InstrumentHandlerDuration provides request duration metrics
type InstrumentHandlerDuration struct {
	Reporter ObserverReporter
	Labeller ServerLabeller

	// Now is the optional strategy for obtaining the system time.  If not supplied, time.Now is used.
	Now func() time.Time
}

func (ihd InstrumentHandlerDuration) Then(next http.Handler) http.Handler {
	if ihd.Reporter == nil {
		return next
	}

	labeller := ihd.Labeller
	if labeller == nil {
		labeller = EmptyLabeller{}
	}

	now := ihd.Now
	if now == nil {
		now = time.Now
	}

	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		start := now()
		next.ServeHTTP(response, request)
		var l xmetrics.Labels
		labeller.ServerLabels(response, request, &l)
		ihd.Reporter.Report(&l, float64(now().Sub(start)))
	})
}

// InstrumentHandlerInFlight records how many current HTTP transactions are being executed by an http.Handler
type InstrumentHandlerInFlight struct {
	Reporter SetterReporter
}

func (ihif InstrumentHandlerInFlight) Then(next http.Handler) http.Handler {
	if ihif.Reporter == nil {
		return next
	}

	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		defer ihif.Reporter.Report(nil, -1.0)
		ihif.Reporter.Report(nil, 1.0)
		next.ServeHTTP(response, request)
	})
}

// RoundTripperFunc is a function type that implements http.RoundTripper
type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (rtf RoundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return rtf(request)
}

// InstrumentRoundTripperCounter provides a simple counting metric for clients executing HTTP transactions
type InstrumentRoundTripperCounter struct {
	Reporter AdderReporter
	Labeller ClientLabeller
}

func (irtc InstrumentRoundTripperCounter) Then(next http.RoundTripper) http.RoundTripper {
	if irtc.Reporter == nil {
		return next
	}

	labeller := irtc.Labeller
	if labeller == nil {
		labeller = EmptyLabeller{}
	}

	return RoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		response, err := next.RoundTrip(request)
		if err == nil {
			var l xmetrics.Labels
			labeller.ClientLabels(response, request, &l)
			irtc.Reporter.Report(&l, 1.0)
		}

		return response, err
	})
}

type InstrumentRoundTripperDuration struct {
	Reporter ObserverReporter
	Labeller ClientLabeller

	// Now is the optional strategy for obtaining the system time.  If not supplied, time.Now is used.
	Now func() time.Time
}

func (irtd InstrumentRoundTripperDuration) Then(next http.RoundTripper) http.RoundTripper {
	if irtd.Reporter == nil {
		return next
	}

	now := irtd.Now
	if now == nil {
		now = time.Now
	}

	labeller := irtd.Labeller
	if labeller == nil {
		labeller = EmptyLabeller{}
	}

	return RoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		start := now()
		response, err := next.RoundTrip(request)
		if err == nil {
			var l xmetrics.Labels
			labeller.ClientLabels(response, request, &l)
			irtd.Reporter.Report(&l, float64(now().Sub(start)))
		}

		return response, err
	})
}

// InstrumentHandlerInFlight provides a gauge of how many in-flight HTTP transactions a client has initiated.
// No labeller is used here, as the reporter must be invoked before the transaction executes to produce a response.
type InstrumentRoundTripperInFlight struct {
	Reporter SetterReporter
}

func (irtif InstrumentRoundTripperInFlight) Then(next http.RoundTripper) http.RoundTripper {
	if irtif.Reporter == nil {
		return next
	}

	return RoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		defer irtif.Reporter.Report(nil, -1.0)
		irtif.Reporter.Report(nil, 1.0)
		return next.RoundTrip(request)
	})
}
