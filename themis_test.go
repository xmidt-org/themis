// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package themis_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/xmidt-org/themis/v2"
	"github.com/xmidt-org/themis/v2/config"
	"github.com/xmidt-org/themis/v2/token"
	"go.uber.org/fx"
)

func Test_themisApp(t *testing.T) {
	t.Run("themis app", func(t *testing.T) {
		require := require.New(t)
		app, err := themis.New(
			fx.Options(
				config.CommandLine{Name: themis.ApplicationName}.Provide(),
				fx.Provide(
					fx.Annotate(func() config.ViperBuilder { return setupViper }, fx.ResultTags(`group:"viperBuilders"`)),
					token.ProvideRemoteClaimsEndpoint,
					func() endpoint.Endpoint { return endpoint.Nop },
				),
			),
		)
		require.NoError(err)

		// only run the program for	a few seconds to make sure it starts
		startCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err = app.Start(startCtx)
		require.NoError(err)
		stopCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err = app.Stop(stopCtx)
		require.NoError(err)
	})
}

func setupViper(in config.ViperIn, v *viper.Viper) error {
	v.SetConfigName(string(in.Name))
	v.AddConfigPath("./cmd/themis")

	return v.ReadInConfig()
}
