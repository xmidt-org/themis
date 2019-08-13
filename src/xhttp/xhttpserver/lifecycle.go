package xhttpserver

import (
	"context"
	"net"
	"xlog"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// OnStart produces a closure that will start the given server appropriately
func OnStart(o Options, s Interface, logger log.Logger, onExit func()) func(context.Context) error {
	return func(ctx context.Context) error {
		l, err := NewListener(o, ctx, net.ListenConfig{})
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
func OnStop(s Interface, logger log.Logger) func(context.Context) error {
	return func(ctx context.Context) error {
		logger.Log(
			level.Key(), level.InfoValue(),
			xlog.MessageKey(), "server stopping",
		)

		return s.Shutdown(ctx)
	}
}
