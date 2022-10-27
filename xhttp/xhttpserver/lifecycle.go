package xhttpserver

import (
	"context"
	"net"

	"github.com/xmidt-org/themis/xlog"
	"go.uber.org/zap"

	"github.com/go-kit/log/level"
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
			logger.Log(
				level.Key(), level.InfoValue(),
				AddressKey(), address,
				xlog.MessageKey(), "starting server",
			)

			err := s.Serve(l)
			logger.Log(
				level.Key(), level.ErrorValue(),
				AddressKey(), address,
				xlog.MessageKey(), "listener exited",
				xlog.ErrorKey(), err,
			)
		}()

		return nil
	}
}

// OnStop produces a closure that will shutdown the server appropriately
func OnStop(s Interface, logger *zap.Logger) func(context.Context) error {
	return func(ctx context.Context) error {
		logger.Log(
			level.Key(), level.InfoValue(),
			xlog.MessageKey(), "server stopping",
		)

		return s.Shutdown(ctx)
	}
}
