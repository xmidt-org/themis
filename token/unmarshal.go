// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package token

import (
	"errors"

	"github.com/go-kit/kit/endpoint"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/xmidt-org/themis/v2/config"
	"github.com/xmidt-org/themis/v2/key"
	"github.com/xmidt-org/themis/v2/random"
	"github.com/xmidt-org/themis/v2/xhttp/xhttpclient"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var (
	ErrRemoteClaimsEndpointMisconfigured = errors.New("remote claims builder must be configured if a remote claim builder endpoint was provided as an fx.Option")
)

func ProvideRemoteClaimsEndpoint(in provideRemoteClaimsEndpointIn) (out provideRemoteClaimsEndpointOut, err error) {
	if in.Options.Remote == nil {
		return provideRemoteClaimsEndpointOut{}, nil
	}

	out.Endpoint, err = newRemoteEndpoint(in.Client, in.Options.Remote)

	return
}

type provideRemoteClaimsEndpointIn struct {
	fx.In

	Client  xhttpclient.Interface `optional:"true"`
	Options Options
}

type provideRemoteClaimsEndpointOut struct {
	fx.Out

	Endpoint endpoint.Endpoint `name:"remote_claims_endpoint"`
}

func ConsumeRemoteClaimsEndpoint(in consumeRemoteClaimsEndpointIn) (out consumeRemoteClaimsEndpointOut, err error) {
	if in.Endpoint != nil && in.Options.Remote == nil {
		return consumeRemoteClaimsEndpointOut{}, ErrRemoteClaimsEndpointMisconfigured
	}

	out.Endpoint = EndpointWrapper(in.Endpoint, in.Options.Remote.URL)

	return
}

type consumeRemoteClaimsEndpointIn struct {
	fx.In

	Endpoint endpoint.Endpoint
	Options  Options
}

type consumeRemoteClaimsEndpointOut struct {
	fx.Out

	Endpoint endpoint.Endpoint `name:"remote_claims_endpoint"`
}

func Unmarshal(configKey string) func(config.Unmarshaller) (Options, error) {
	return func(unmarshaller config.Unmarshaller) (o Options, err error) {
		return o, unmarshaller.UnmarshalKey(configKey, &o)
	}
}

type TokenIn struct {
	fx.In

	Logger         *zap.Logger
	Noncer         random.Noncer `optional:"true"`
	Keys           key.Registry
	Options        Options
	RemoteEndpoint endpoint.Endpoint        `name:"remote_claims_endpoint"`
	TrustCounter   *prometheus.CounterVec   `name:"trust_total"`
	RemoteResults  *prometheus.CounterVec   `name:"remote_claims_api_result_total"`
	RemoteDuration *prometheus.HistogramVec `name:"remote_claims_api_request_duration_seconds"`
}

type TokenOut struct {
	fx.Out

	ClaimBuilder  ClaimBuilder
	Factory       Factory
	IssueHandler  IssueHandler
	ClaimsHandler ClaimsHandler
}

// TokenFactory returns an uber/fx style factory that produces the relevant components for
// a single token factory.
func TokenFactory(b ...RequestBuilder) func(TokenIn) (TokenOut, error) {
	return func(in TokenIn) (TokenOut, error) {
		if in.Options.ClientCertificates != nil {
			in.Logger.Info("trust settings", zap.Any("trust_config", in.Options.ClientCertificates.Trust))
		} else {
			in.Logger.Info("trust settings", zap.Any("trust_config", Trust{}.enforceDefaults()))
		}

		cb, err := NewClaimBuilders(in.Noncer, in.RemoteEndpoint, in.Options, in.TrustCounter, in.RemoteResults, in.RemoteDuration)
		if err != nil {
			return TokenOut{}, err
		}

		f, err := NewFactory(in.Options, cb, in.Keys)
		if err != nil {
			return TokenOut{}, err
		}

		rb, err := NewRequestBuilders(in.Options)
		if err != nil {
			return TokenOut{}, err
		}

		rb = append(rb, b...)
		return TokenOut{
			ClaimBuilder: cb,
			Factory:      f,
			IssueHandler: NewIssueHandler(
				NewIssueEndpoint(f),
				rb,
			),
			ClaimsHandler: NewClaimsHandler(
				NewClaimsEndpoint(cb),
				rb,
			),
		}, nil
	}
}
