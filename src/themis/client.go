package main

import (
	"xhttp/xhttpclient"

	"go.uber.org/fx"
)

type ClientChainIn struct {
	fx.In
	ClientMetricsIn
}

// provideClientChain provides the global decoration for all HTTP clients
func provideClientChain(in ClientChainIn) xhttpclient.Chain {
	return xhttpclient.NewChain(
		in.RequestCount.Then,
		in.RequestDuration.Then,
		in.RequestsInFlight.Then,
	)
}
