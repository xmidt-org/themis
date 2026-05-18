// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

func TestMain(t *testing.T) {
	originalArgs := os.Args
	require := require.New(t)

	// FailureTimeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	sig := app(ctx)
	require.Equal(syscall.SIGTRAP, sig, sig)
	cancel()

	// FailureBadArgs
	os.Args = append(os.Args, "--THisShoUldNotEXIST")
	sig = app(context.Background())
	require.Equal(syscall.SIGINT, sig, sig)
	os.Args = originalArgs

	// HelpSuccess
	os.Args = append(os.Args, "--help")
	sig = app(context.Background())
	require.Zero(int(sig), sig)
	os.Args = originalArgs
	cancel()

	// Success
	sig = app(context.Background(), fx.Invoke(
		func(lifecycle fx.Lifecycle, shutdown fx.Shutdowner) {
			lifecycle.Append(fx.Hook{
				OnStart: func(ctx context.Context) error { return shutdown.Shutdown(fx.ExitCode(0)) },
			})
		},
	))
	require.Zero(int(sig), sig)
}
