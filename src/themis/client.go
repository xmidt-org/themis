package main

import (
	"xhttp/xhttpclient"
	"xmetrics/xmetricshttp"

	"go.uber.org/fx"
)

type ClientIn struct {
	fx.In
	RequestCount     xmetricshttp.RoundTripperCounter  `name:"client_request_count"`
	RequestDuration  xmetricshttp.RoundTripperDuration `name:"client_request_duration_ms"`
	RequestsInFlight xmetricshttp.RoundTripperInFlight `name:"client_requests_in_flight"`
}

// provideClientChain provides the global decoration for all HTTP clients
func provideClientChain(in ClientIn) xhttpclient.Chain {
	return xhttpclient.NewChain(
		in.RequestCount.Then,
		in.RequestDuration.Then,
		in.RequestsInFlight.Then,
	)
}
