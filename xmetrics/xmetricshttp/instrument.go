package xmetricshttp

import (
	"net/http"
	"time"

	"github.com/xmidt-org/themis/xhttp/xhttpclient"
	"github.com/xmidt-org/themis/xmetrics"
)

// HandlerCounter provides a simple count metric of HTTP transactions
type HandlerCounter struct {
	Metric   xmetrics.Adder
	Labeller ServerLabeller
}

func (ihc HandlerCounter) Then(next http.Handler) http.Handler {
	if ihc.Metric == nil {
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
		ihc.Metric.Add(&l, 1.0)
	})
}

// HandlerDuration provides request duration metrics
type HandlerDuration struct {
	Metric   xmetrics.Observer
	Labeller ServerLabeller

	// Now is the optional strategy for obtaining the system time.  If not supplied, time.Now is used.
	Now func() time.Time

	// Units is the time unit to report the metric in.  If unset, time.Millisecond is used.  Any of the
	// time duration constants can be used here, e.g. time.Second or time.Minute.
	Units time.Duration
}

func (ihd HandlerDuration) Then(next http.Handler) http.Handler {
	if ihd.Metric == nil {
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

	units := ihd.Units
	if units <= 0 {
		units = time.Millisecond
	}

	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		start := now()
		next.ServeHTTP(response, request)
		var l xmetrics.Labels
		labeller.ServerLabels(response, request, &l)
		ihd.Metric.Observe(
			&l,
			float64(now().Sub(start)/units),
		)
	})
}

// HandlerInFlight records how many current HTTP transactions are being executed by an http.Handler
type HandlerInFlight struct {
	Metric xmetrics.GaugeAdder
}

func (ihif HandlerInFlight) Then(next http.Handler) http.Handler {
	if ihif.Metric == nil {
		return next
	}

	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		defer ihif.Metric.GaugeAdd(nil, -1.0)
		ihif.Metric.GaugeAdd(nil, 1.0)
		next.ServeHTTP(response, request)
	})
}

// RoundTripperCounter provides a simple counting metric for clients executing HTTP transactions
type RoundTripperCounter struct {
	Metric   xmetrics.Adder
	Labeller ClientLabeller
}

func (irtc RoundTripperCounter) Then(next http.RoundTripper) http.RoundTripper {
	if irtc.Metric == nil {
		return next
	}

	labeller := irtc.Labeller
	if labeller == nil {
		labeller = EmptyLabeller{}
	}

	return xhttpclient.RoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		response, err := next.RoundTrip(request)
		if err == nil {
			var l xmetrics.Labels
			labeller.ClientLabels(response, request, &l)
			irtc.Metric.Add(&l, 1.0)
		}

		return response, err
	})
}

type RoundTripperDuration struct {
	Metric   xmetrics.Observer
	Labeller ClientLabeller

	// Now is the optional strategy for obtaining the system time.  If not supplied, time.Now is used.
	Now func() time.Time

	// Units is the time unit to report the metric in.  If unset, time.Millisecond is used.  Any of the
	// time duration constants can be used here, e.g. time.Second or time.Minute.
	Units time.Duration
}

func (irtd RoundTripperDuration) Then(next http.RoundTripper) http.RoundTripper {
	if irtd.Metric == nil {
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

	units := irtd.Units
	if units <= 0 {
		units = time.Millisecond
	}

	return xhttpclient.RoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		start := now()
		response, err := next.RoundTrip(request)
		if err == nil {
			var l xmetrics.Labels
			labeller.ClientLabels(response, request, &l)
			irtd.Metric.Observe(
				&l,
				float64(now().Sub(start)/units),
			)
		}

		return response, err
	})
}

// HandlerInFlight provides a gauge of how many in-flight HTTP transactions a client has initiated.
// No labeller is used here, as the reporter must be invoked before the transaction executes to produce a response.
type RoundTripperInFlight struct {
	Metric xmetrics.GaugeAdder
}

func (irtif RoundTripperInFlight) Then(next http.RoundTripper) http.RoundTripper {
	if irtif.Metric == nil {
		return next
	}

	return xhttpclient.RoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		defer irtif.Metric.GaugeAdd(nil, -1.0)
		irtif.Metric.GaugeAdd(nil, 1.0)
		return next.RoundTrip(request)
	})
}
