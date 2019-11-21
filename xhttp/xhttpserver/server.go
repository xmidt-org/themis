package xhttpserver

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/xmidt-org/themis/xlog/xloghttp"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
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
func NewServerChain(o Options, l log.Logger, pb ...xloghttp.ParameterBuilder) alice.Chain {
	chain := alice.New(
		ResponseHeaders{Header: o.Header}.Then,
		Busy{MaxConcurrentRequests: o.MaxConcurrentRequests}.Then,
	)

	if !o.DisableTracking {
		chain = chain.Append(UseTrackingWriter)
	}

	if !o.DisableHandlerLogger {
		chain = chain.Append(
			xloghttp.Logging{Base: l, Builders: pb}.Then,
		)
	}

	return chain
}

// New constructs a basic HTTP server instance.  The supplied logger is enriched with information
// about the server and returned for use by higher-level code.
func New(o Options, l log.Logger, h http.Handler) Interface {
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

		ErrorLog: xloghttp.NewErrorLog(
			o.Address,
			log.WithPrefix(l, level.Key(), level.ErrorValue()),
		),
	}

	if o.LogConnectionState {
		s.ConnState = xloghttp.NewConnStateLogger(
			l,
			"connState",
			level.DebugValue(),
		)
	}

	if o.DisableHTTPKeepAlives {
		s.SetKeepAlivesEnabled(false)
	}

	return s
}
