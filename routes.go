// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"errors"

	"github.com/xmidt-org/themis/key"
	"github.com/xmidt-org/themis/token"
	"github.com/xmidt-org/themis/xhealth"
	"github.com/xmidt-org/themis/xhttp/xhttpserver"
	"github.com/xmidt-org/themis/xhttp/xhttpserver/pprof"
	"github.com/xmidt-org/themis/xmetrics"
	"github.com/xmidt-org/themis/xmetrics/xmetricshttp"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
)

type ServerChainIn struct {
	fx.In

	RequestCount     *prometheus.CounterVec   `name:"server_request_count"`
	RequestDuration  *prometheus.HistogramVec `name:"server_request_duration_ms"`
	RequestsInFlight *prometheus.GaugeVec     `name:"server_requests_in_flight"`
}

func provideServerChainFactory(in ServerChainIn) xhttpserver.ChainFactory {
	return xhttpserver.ChainFactoryFunc(func(name string, o xhttpserver.Options) (alice.Chain, error) {
		var (
			curryLabel = prometheus.Labels{
				ServerLabel: name,
			}

			serverLabellers = xmetricshttp.NewServerLabellers(
				xmetricshttp.CodeLabeller{},
				xmetricshttp.MethodLabeller{},
			)
		)

		requestCount, err := in.RequestCount.CurryWith(curryLabel)
		if err != nil {
			return alice.Chain{}, err
		}

		requestDuration, err := in.RequestDuration.CurryWith(curryLabel)
		if err != nil {
			return alice.Chain{}, err
		}

		requestsInFlight, err := in.RequestsInFlight.CurryWith(curryLabel)
		if err != nil {
			return alice.Chain{}, err
		}

		return alice.New(
			xmetricshttp.HandlerCounter{
				Metric:   xmetrics.LabelledCounterVec{CounterVec: requestCount},
				Labeller: serverLabellers,
			}.Then,
			xmetricshttp.HandlerDuration{
				Metric:   xmetrics.LabelledObserverVec{ObserverVec: requestDuration},
				Labeller: serverLabellers,
			}.Then,
			xmetricshttp.HandlerInFlight{
				Metric: xmetrics.LabelledGaugeVec{GaugeVec: requestsInFlight},
			}.Then,
		), nil
	})
}

type KeyRoutesIn struct {
	fx.In
	Router     *mux.Router `name:"servers.key"`
	Handler    key.Handler
	HandlerJWK key.HandlerJWK
}

func BuildKeyRoutes(in KeyRoutesIn) {
	if in.Router != nil {
		keys := in.Router.PathPrefix("/keys/{kid}").Methods("GET").Subrouter()

		keys.Headers("Accept", key.ContentTypePEM).Handler(in.Handler)
		keys.Headers("Accept", key.ContentTypeJWK).Handler(in.HandlerJWK)
		keys.Path("").Handler(in.Handler) // default
		keys.Path("/key.pem").Handler(in.Handler)
		keys.Path("/key.json").Handler(in.HandlerJWK)
	}
}

type IssuerRoutesIn struct {
	fx.In
	Router  *mux.Router `name:"servers.issuer"`
	Handler token.IssueHandler
}

func BuildIssuerRoutes(in IssuerRoutesIn) {
	if in.Router != nil && in.Handler != nil {
		in.Router.Handle("/issue", in.Handler).Methods("GET")
	}
}

type ClaimsRoutesIn struct {
	fx.In
	Router  *mux.Router `name:"servers.claims"`
	Handler token.ClaimsHandler
}

func BuildClaimsRoutes(in ClaimsRoutesIn) {
	if in.Router != nil && in.Handler != nil {
		in.Router.Handle("/claims", in.Handler).Methods("GET")
	}
}

// CheckServerRequirements is an fx.Invoke function that does post-configuration verification
// that we have required servers.  The valid server configurations are:
//
//	Both keys and issuer present.  Claims is optional in this case
//	Neither keys or issuer present.  Claims is required in this case
//
// Any other arrangements results in an error.
func CheckServerRequirements(k KeyRoutesIn, i IssuerRoutesIn, c ClaimsRoutesIn) error {
	if k.Router != nil && i.Router != nil {
		// all good ... no need to check anything else
		return nil
	}

	if k.Router == nil && i.Router == nil {
		if c.Router == nil {
			return errors.New("a claims server is required if no keys or issuer server is configured")
		}

		// Only a claims server is allowed
		return nil
	}

	if k.Router != nil {
		return errors.New("if a keys server is configured, an issuer server must be configured")
	}

	if i.Router != nil {
		return errors.New("if an issuer server is configured, a keys server must be configured")
	}

	return nil
}

type MetricsRoutesIn struct {
	fx.In
	Router  *mux.Router `name:"servers.metrics"`
	Handler xmetricshttp.Handler
}

func BuildMetricsRoutes(in MetricsRoutesIn) {
	if in.Router != nil && in.Handler != nil {
		in.Router.Handle("/metrics", in.Handler).Methods("GET")
	}
}

type HealthRoutesIn struct {
	fx.In
	Router  *mux.Router `name:"servers.health"`
	Handler xhealth.Handler
}

func BuildHealthRoutes(in HealthRoutesIn) {
	if in.Router != nil && in.Handler != nil {
		in.Router.Handle("/health", in.Handler).Methods("GET")
	}
}

type PprofRoutesIn struct {
	fx.In
	Router *mux.Router `name:"servers.pprof"`
}

func BuildPprofRoutes(in PprofRoutesIn) {
	if in.Router != nil {
		pprof.BuildRoutes(in.Router)
	}
}
