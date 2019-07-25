package main

import (
	"xhttp/xhttpclient"

	"go.uber.org/fx"
)

type ClientIn struct {
	fx.In
	ClientInstrumentsIn
	xhttpclient.UnmarshalIn
}

func provideClient(configKey string) func(ClientIn) (xhttpclient.Interface, error) {
	return func(in ClientIn) (xhttpclient.Interface, error) {
		c, err := xhttpclient.Unmarshal(
			configKey,
			in.RequestCount.Then,
			in.RequestDuration.Then,
			in.RequestsInFlight.Then,
		)(in.UnmarshalIn)

		if err != nil {
			return nil, err
		}

		return c, nil
	}
}
