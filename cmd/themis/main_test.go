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
	done := make(chan struct{})
	require := require.New(t)

	// StartTimeout
	ctx, cancel := context.WithTimeout(t.Context(), time.Nanosecond)
	sig := app(done, ctx, t.Context())
	cancel()
	require.Equal(syscall.SIGTRAP, sig, sig)

	// FailureBadArgs
	os.Args = append(os.Args, "--THisShoUldNotEXIST")
	sig = app(done, t.Context(), t.Context())
	require.Equal(syscall.SIGINT, sig, sig)
	os.Args = originalArgs

	// FileFailure
	os.Args = append(os.Args, "--file", "some_name")
	sig = app(done, t.Context(), t.Context())
	require.Equal(syscall.SIGINT, sig, sig)
	os.Args = originalArgs

	// HelpSuccess
	os.Args = append(os.Args, "--help")
	sig = app(done, t.Context(), t.Context())
	require.Zero(int(sig), sig)
	os.Args = originalArgs

	// VersionSuccess
	os.Args = append(os.Args, "--version")
	sig = app(done, t.Context(), t.Context())
	require.Zero(int(sig), sig)
	os.Args = originalArgs

	// Trigger graceful shutdown attempts.
	close(done)

	// Bad Shutdown hook
	ctx, cancel = context.WithTimeout(t.Context(), time.Nanosecond)
	sig = app(done, t.Context(), ctx, fx.Invoke(
		func(lifecycle fx.Lifecycle, shutdown fx.Shutdowner) {
			lifecycle.Append(fx.Hook{
				OnStop: func(context.Context) error {
					shutdown.Shutdown()

					return context.DeadlineExceeded
				},
			})
		},
	))
	cancel()
	require.Equal(syscall.SIGHUP, sig, sig)

	// StopTimeout Success
	ctx, cancel = context.WithTimeout(t.Context(), time.Nanosecond)
	sig = app(done, t.Context(), ctx)
	cancel()
	require.Zero(int(sig), sig)

	// DevSuccess
	os.Args = append(os.Args, "--dev")
	sig = app(done, t.Context(), t.Context())
	require.Zero(int(sig), sig)
	os.Args = originalArgs

	// DebugSuccess
	os.Args = append(os.Args, "--debug")
	sig = app(done, t.Context(), t.Context())
	require.Zero(int(sig), sig)
	os.Args = originalArgs

	// FileSuccess
	os.Args = append(os.Args, "--file", "./themis.yaml")
	sig = app(done, t.Context(), t.Context())
	require.Zero(int(sig), sig)
	os.Args = originalArgs

	// FileSuccess
	os.Args = append(os.Args, "--iss", "123")
	sig = app(done, t.Context(), t.Context())
	require.Zero(int(sig), sig)
	os.Args = originalArgs

	// Success
	sig = app(done, t.Context(), t.Context())
	require.Zero(int(sig), sig)
}
