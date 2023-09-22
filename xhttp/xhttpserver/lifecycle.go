// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package xhttpserver

import (
	"context"
	"net"

	"go.uber.org/zap"
)

// OnStart produces a closure that will start the given server appropriately
func OnStart(o Options, s Interface, logger *zap.Logger, onExit func()) func(context.Context) error {
	return func(ctx context.Context) error {
		tcfg, err := NewTlsConfig(o.Tls)
		if err != nil {
			return err
		}

		l, err := NewListener(ctx, o, net.ListenConfig{}, tcfg)
		if err != nil {
			return err
		}

		go func() {
			if onExit != nil {
				defer onExit()
			}

			address := l.Addr().String()
			logger.Info("starting server", zap.String(AddressKey(), address))

			err := s.Serve(l)
			logger.Error("listener exited", zap.String(AddressKey(), address), zap.Error(err))
		}()

		return nil
	}
}

// OnStop produces a closure that will shutdown the server appropriately
func OnStop(s Interface, logger *zap.Logger) func(context.Context) error {
	return func(ctx context.Context) error {
		logger.Info("server stopping")

		return s.Shutdown(ctx)
	}
}
