// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package token

import (
	"errors"

	"github.com/go-kit/kit/endpoint"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/xmidt-org/themis/config"
	"github.com/xmidt-org/themis/key"
	"github.com/xmidt-org/themis/random"
	"github.com/xmidt-org/themis/xhttp/xhttpclient"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var (
	ErrRemoteClaimsEndpointConflict = errors.New("remote claims builder's gokit endpoint can either be configured or provided as an fx.Option")
)

type RemoteClaimsEndpointIn struct {
	fx.In

	Endpoint endpoint.Endpoint     `optional:"true"`
	Client   xhttpclient.Interface `optional:"true"`
	Options  Options
}

type RemoteClaimsEndpointOut struct {
	fx.Out

	Endpoint endpoint.Endpoint `name:"remote_claims_endpoint"`
}

func RemoteClaimsEndpoint(in RemoteClaimsEndpointIn) (RemoteClaimsEndpointOut, error) {
	if in.Client != nil && in.Endpoint != nil {
		return RemoteClaimsEndpointOut{}, ErrRemoteClaimsEndpointConflict
	}

	var (
		endpoint endpoint.Endpoint
		err      error
	)
	if in.Endpoint != nil {
		endpoint = in.Endpoint
	} else if in.Options.Remote != nil {
		endpoint, err = newRemoteEndpoint(in.Client, in.Options.Remote)
	}

	return RemoteClaimsEndpointOut{Endpoint: endpoint}, err
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
	RemoteEndpoint endpoint.Endpoint        `name:"remote_claims_endpoint" optional:"true"`
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
			in.Logger.Info("trust settings", zap.Reflect("trust", in.Options.ClientCertificates.Trust))
		} else {
			in.Logger.Info("trust settings", zap.Reflect("trust", Trust{}.enforceDefaults()))
		}

		cb, err := NewClaimBuilders(in.Noncer, in.RemoteEndpoint, in.Options, in.RemoteResults, in.RemoteDuration)
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
