// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package xhttpserver

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/xmidt-org/sallust"
	"github.com/xmidt-org/sallust/sallusthttp"
	"go.uber.org/zap"

	"github.com/justinas/alice"
)

// Interface is the expected behavior of a server
type Interface interface {
	// Serve handles both TLS and non-TLS listeners.  Code in this package takes care of creating
	// a *tls.Config and TLS listener.
	Serve(l net.Listener) error

	// Shutdown gracefully shuts down the server
	Shutdown(context.Context) error
}

// Options represent the configurable options for creating a server, typically unmarshalled from an
// external source.
type Options struct {
	Address string
	Network string
	Tls     *Tls

	LogConnectionState    bool
	DisableHTTPKeepAlives bool
	MaxHeaderBytes        int

	IdleTimeout           time.Duration
	ReadHeaderTimeout     time.Duration
	ReadTimeout           time.Duration
	WriteTimeout          time.Duration
	MaxConcurrentRequests int

	DisableTCPKeepAlives bool
	TCPKeepAlivePeriod   time.Duration

	Header               http.Header
	DisableTracking      bool
	DisableHandlerLogger bool
}

// NewServerChain produces the standard constructor chain for a server, primarily using configuration.
func NewServerChain(o Options, l *zap.Logger, fbs ...sallusthttp.FieldBuilder) alice.Chain {
	bs := sallusthttp.Builders{}
	chain := alice.New(
		ResponseHeaders{Header: o.Header}.Then,
		Busy{MaxConcurrentRequests: o.MaxConcurrentRequests}.Then,
	)

	bs.AddFields(fbs...)
	if !o.DisableTracking {
		chain = chain.Append(UseTrackingWriter)
	}

	if !o.DisableHandlerLogger {
		chain = chain.Append(
			func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
					requestLogger := bs.Build(request, l)
					requestLogger.Info(
						"tls info",
						connectionStateField("state", request.TLS),
					)

					next.ServeHTTP(
						response,
						sallusthttp.With(request, requestLogger),
					)
				})
			},
		)
	}

	return chain
}

// New constructs a basic HTTP server instance.  The supplied logger is enriched with information
// about the server and returned for use by higher-level code.
func New(o Options, l *zap.Logger, h http.Handler) Interface {
	s := &http.Server{
		// we don't need this technically, because we create a listener
		// it's here for other code to inspect
		Addr:    o.Address,
		Handler: h,

		MaxHeaderBytes:    o.MaxHeaderBytes,
		IdleTimeout:       o.IdleTimeout,
		ReadHeaderTimeout: o.ReadHeaderTimeout,
		ReadTimeout:       o.ReadTimeout,
		WriteTimeout:      o.WriteTimeout,

		ErrorLog: sallust.NewServerLogger(
			o.Address,
			l,
		),
	}

	if o.LogConnectionState {
		s.ConnState = sallusthttp.NewConnStateLogger(
			l,
			zap.DebugLevel,
		)
	}

	if o.DisableHTTPKeepAlives {
		s.SetKeepAlivesEnabled(false)
	}

	return s
}
