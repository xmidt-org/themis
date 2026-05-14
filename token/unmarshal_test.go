// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package token

import (
	"testing"

	"github.com/go-kit/kit/endpoint"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xmidt-org/sallust"
	"github.com/xmidt-org/themis/config"
	"github.com/xmidt-org/themis/key"
	"github.com/xmidt-org/themis/xhttp/xhttpclient"
	"github.com/xmidt-org/themis/xmetrics/xmetricshttp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func testUnmarshalError(t *testing.T) {
	var (
		assert  = assert.New(t)
		factory Factory
	)

	app := fx.New(
		fx.NopLogger,
		fx.Provide(
			config.ProvideViper,
			fx.Annotate(config.Json(`
					{
						"token": {
							"nonce": "this is not a valid bool"
						}
					}
				`), fx.ResultTags(`group:"viperBuilders"`)),
			func() key.Registry { return key.NewRegistry(nil) },
			Unmarshal("token"),
		),
		fx.Populate(&factory),
	)

	assert.Error(app.Err())
	assert.Nil(factory)
}

func testUnmarshalClaimBuilderError(t *testing.T) {
	var (
		assert  = assert.New(t)
		factory Factory

		app = fx.New(
			fx.NopLogger,
			fx.Provide(
				config.ProvideViper,
				fx.Annotate(config.Json(`
						{
							"token": {
								"metadata": [
									{
										"key": "bad"
									}
								],
								"remote": {
									"url": "http://foobar.com"
								}
							}
						}
					`), fx.ResultTags(`group:"viperBuilders"`)),
				func() key.Registry { return key.NewRegistry(nil) },
				Unmarshal("token"),
			),
			fx.Populate(&factory),
		)
	)

	assert.Error(app.Err())
	assert.Nil(factory)
}

func testUnmarshalFactoryError(t *testing.T) {
	var (
		assert  = assert.New(t)
		factory Factory

		app = fx.New(
			fx.NopLogger,
			fx.Provide(
				config.ProvideViper,
				fx.Annotate(config.Json(`
						{
							"token": {
								"alg": "this is not a signing method"
							}
						}
					`), fx.ResultTags(`group:"viperBuilders"`)),
				func() key.Registry { return key.NewRegistry(nil) },
				Unmarshal("token"),
			),
			fx.Populate(&factory),
		)
	)

	assert.Error(app.Err())
	assert.Nil(factory)
}

func testUnmarshalRequestBuilderError(t *testing.T) {
	var (
		assert  = assert.New(t)
		factory Factory

		app = fx.New(
			fx.NopLogger,
			fx.Provide(
				config.ProvideViper,
				fx.Annotate(config.Json(`
						{
							"token": {
								"claims": [
									{
										"key": "bad",
										"header": "X-Bad",
										"parameter": "bad",
										"variable": "bad"
									}
								]
							}
						}
					`), fx.ResultTags(`group:"viperBuilders"`)),
				func() key.Registry { return key.NewRegistry(nil) },
				Unmarshal("token"),
			),
			fx.Populate(&factory),
		)
	)

	assert.Error(app.Err())
	assert.Nil(factory)
}

func testUnmarshalRemoteEndpointMisconfigured(t *testing.T) {
	type requiredIn struct {
		fx.In

		Endpoint endpoint.Endpoint `name:"remote_claims_endpoint"`
		Client   xhttpclient.Interface
	}

	var (
		assert  = assert.New(t)
		require = require.New(t)
		factory Factory
		app     = fx.New(
			ProvideMetrics(),
			fx.Provide(
				sallust.Default,
				config.ProvideViper,
				fx.Annotate(config.Json(`
						{
							"prometheus": {
								"defaultNamespace": "xmidt",
								"defaultSubsystem": "issuer",
								"constLabels":{
									"development": "true"
								}
							},
							"token": {
								"claims": [
									{
										"key": "static",
										"value": "foo"
									}
								]
							},
						}
					`), fx.ResultTags(`group:"viperBuilders"`)),
				func() key.Registry { return key.NewRegistry(nil) },
				Unmarshal("token"),
				xhttpclient.Unmarshal{Key: "client"}.Provide,
				TokenFactory(),
				RemoteClaimsEndpoint,
				xmetricshttp.Unmarshal("prometheus", promhttp.HandlerOpts{}),
				func() endpoint.Endpoint { return endpoint.Nop },
			),
			fx.Populate(&factory),
			fx.Invoke(func(in requiredIn) {
				require.NotNil(in.Endpoint)
				require.NotNil(in.Client)
			}),
		)
	)

	assert.Error(app.Err())
	assert.Nil(factory)
}

func testUnmarshalWithoutRemoteEndpointSuccess(t *testing.T) {
	type requiredIn struct {
		fx.In

		Endpoint endpoint.Endpoint     `name:"remote_claims_endpoint" optional:"true"`
		Client   xhttpclient.Interface `optional:"true"`
	}

	var (
		assert  = assert.New(t)
		require = require.New(t)
		factory Factory

		app = fxtest.New(t,

			ProvideMetrics(),
			fx.Provide(
				sallust.Default,
				config.ProvideViper,
				fx.Annotate(func() config.ViperBuilder {
					return config.Json(`
						{
							"prometheus": {
								"defaultNamespace": "xmidt",
								"defaultSubsystem": "issuer",
								"constLabels":{
									"development": "true"
								}
							},
							"token": {
								"claims": [
									{
										"key": "static",
										"value": "foo"
									}
								]
							}
						}
					`)
				}, fx.ResultTags(`group:"viperBuilders"`)),
				func() key.Registry { return key.NewRegistry(nil) },
				Unmarshal("token"),
				xhttpclient.Unmarshal{Key: "client"}.Provide,
				TokenFactory(),
				RemoteClaimsEndpoint,
				xmetricshttp.Unmarshal("prometheus", promhttp.HandlerOpts{}),
			),
			fx.Populate(&factory),
			fx.Invoke(func(in requiredIn) {
				require.Nil(in.Endpoint)
				require.Nil(in.Client)
			}),
		)
	)
	assert.NoError(app.Err())
	assert.NotNil(factory)
}

func testUnmarshalWithProvidedRemoteEndpointSuccess(t *testing.T) {
	type requiredIn struct {
		fx.In

		Endpoint endpoint.Endpoint     `name:"remote_claims_endpoint"`
		Client   xhttpclient.Interface `optional:"true"`
	}

	var (
		assert  = assert.New(t)
		require = require.New(t)
		factory Factory
		app     = fxtest.New(t,
			ProvideMetrics(),
			fx.Provide(
				sallust.Default,
				config.ProvideViper,
				fx.Annotate(func() config.ViperBuilder {
					return config.Json(`
						{
							"prometheus": {
								"defaultNamespace": "xmidt",
								"defaultSubsystem": "issuer",
								"constLabels":{
									"development": "true"
								}
							},
							"token": {
								"claims": [
									{
										"key": "static",
										"value": "foo"
									}
								],
								"remote": {
									"method": "post",
									"url": "https//example.com"
								}
							}
						}
					`)
				}, fx.ResultTags(`group:"viperBuilders"`)),
				func() key.Registry { return key.NewRegistry(nil) },
				Unmarshal("token"),
				xhttpclient.Unmarshal{Key: "client"}.Provide,
				TokenFactory(),
				RemoteClaimsEndpoint,
				xmetricshttp.Unmarshal("prometheus", promhttp.HandlerOpts{}),
				func() endpoint.Endpoint { return endpoint.Nop },
			),
			fx.Populate(&factory),
			fx.Invoke(func(in requiredIn) {
				require.NotNil(in.Endpoint)
				require.Nil(in.Client)
			}),
		)
	)
	assert.NoError(app.Err())
	assert.NotNil(factory)
}

func testUnmarshalWithConfiguredRemoteEndpointSuccess(t *testing.T) {
	type requiredIn struct {
		fx.In

		Endpoint endpoint.Endpoint     `name:"remote_claims_endpoint"`
		Client   xhttpclient.Interface `optional:"true"`
	}

	var (
		assert  = assert.New(t)
		require = require.New(t)
		factory Factory
		app     = fxtest.New(t,
			ProvideMetrics(),
			fx.Provide(
				sallust.Default,
				config.ProvideViper,
				fx.Annotate(func() config.ViperBuilder {
					return config.Json(`
						{
							"prometheus": {
								"defaultNamespace": "xmidt",
								"defaultSubsystem": "issuer",
								"constLabels":{
									"development": "true"
								}
							},
							"token": {
								"claims": [
									{
										"key": "static",
										"value": "foo"
									}
								],
								"remote": {
									"method": "post",
									"url": "https//example.com"
								}
							}
						}
					`)
				}, fx.ResultTags(`group:"viperBuilders"`)),
				func() key.Registry { return key.NewRegistry(nil) },
				Unmarshal("token"),
				xhttpclient.Unmarshal{Key: "client"}.Provide,
				TokenFactory(),
				RemoteClaimsEndpoint,
				xmetricshttp.Unmarshal("prometheus", promhttp.HandlerOpts{}),
			),
			fx.Populate(&factory),
			fx.Invoke(func(in requiredIn) {
				require.NotNil(in.Endpoint)
				require.Nil(in.Client)
			}),
		)
	)
	assert.NoError(app.Err())
	assert.NotNil(factory)
}

func testUnmarshalWithConfiguredRemoteEndpointAndClientSuccess(t *testing.T) {
	type requiredIn struct {
		fx.In

		Endpoint endpoint.Endpoint     `name:"remote_claims_endpoint"`
		Client   xhttpclient.Interface `optional:"true"`
	}

	var (
		assert  = assert.New(t)
		require = require.New(t)
		factory Factory
		app     = fxtest.New(t,
			ProvideMetrics(),
			fx.Provide(
				sallust.Default,
				config.ProvideViper,
				fx.Annotate(func() config.ViperBuilder {
					return config.Json(`
						{
							"prometheus": {
								"defaultNamespace": "xmidt",
								"defaultSubsystem": "issuer",
								"constLabels":{
									"development": "true"
								}
							},
							"token": {
								"claims": [
									{
										"key": "static",
										"value": "foo"
									}
								],
								"remote": {
									"method": "post",
									"url": "https//example.com"
								}
							},
							"client": {}
						}
					`)
				}, fx.ResultTags(`group:"viperBuilders"`)),
				func() key.Registry { return key.NewRegistry(nil) },
				Unmarshal("token"),
				xhttpclient.Unmarshal{Key: "client"}.Provide,
				TokenFactory(),
				RemoteClaimsEndpoint,
				xmetricshttp.Unmarshal("prometheus", promhttp.HandlerOpts{}),
			),
			fx.Populate(&factory),
			fx.Invoke(func(in requiredIn) {
				require.NotNil(in.Endpoint)
				require.NotNil(in.Client)
			}),
		)
	)
	assert.NoError(app.Err())
	assert.NotNil(factory)
}

func TestUnmarshal(t *testing.T) {
	t.Run("Error", testUnmarshalError)
	t.Run("ClaimBuilderError", testUnmarshalClaimBuilderError)
	t.Run("FactoryError", testUnmarshalFactoryError)
	t.Run("RequestBuilderError", testUnmarshalRequestBuilderError)
	t.Run("RemoteEndpointConflictMisconfigured", testUnmarshalRemoteEndpointMisconfigured)
	t.Run("UnmarshalWithoutRemoteEndpointSuccess", testUnmarshalWithoutRemoteEndpointSuccess)
	t.Run("UnmarshalWithProvidedRemoteEndpointSuccess", testUnmarshalWithProvidedRemoteEndpointSuccess)
	t.Run("UnmarshalWithConfiguredRemoteEndpointSuccess", testUnmarshalWithConfiguredRemoteEndpointSuccess)
	t.Run("UnmarshalWithConfiguredRemoteEndpointAndClientSuccess", testUnmarshalWithConfiguredRemoteEndpointAndClientSuccess)
}
