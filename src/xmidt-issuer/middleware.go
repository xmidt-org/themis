package main

import (
	"xhttp/xhttpserver"
	"xmetrics/xmetricshttp"

	"github.com/gorilla/mux"
	"go.uber.org/fx"
)

type MiddlewareIn struct {
	fx.In

	ResponseHeaders xhttpserver.ResponseHeaders

	RequestCount    xmetricshttp.InstrumentHandlerCounter
	RequestDuration xmetricshttp.InstrumentHandlerDuration
	InFlight        xmetricshttp.InstrumentHandlerInFlight
}

func provideMiddleware(in MiddlewareIn) []mux.MiddlewareFunc {
	return []mux.MiddlewareFunc{
		xhttpserver.TrackWriter,
		in.RequestCount.Then,
		in.RequestDuration.Then,
		in.InFlight.Then,
		in.ResponseHeaders.Then,
	}
}
