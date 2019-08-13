package xhttpserver

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"time"
	"xlog/xloghttp"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/justinas/alice"
)

const (
	defaultTCPKeepAlivePeriod time.Duration = 3 * time.Minute // the value used internally by net/http
)

var (
	ErrNoAddress                      = errors.New("A server bind address must be specified")
	ErrUnableToAddClientCACertificate = errors.New("Unable to add client CA certificate")
)

type Options struct {
	Name    string
	Address string
	Tls     *Tls

	LogConnectionState    bool
	DisableHTTPKeepAlives bool
	MaxHeaderBytes        int

	IdleTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration

	DisableTCPKeepAlives bool
	TCPKeepAlivePeriod   time.Duration

	Header               http.Header
	DisableTracking      bool
	DisableHandlerLogger bool
	DisableParseForm     bool
}

// Interface is the expected behavior of a server
type Interface interface {
	Serve(l net.Listener) error
	ServeTLS(l net.Listener, cert, key string) error
	Shutdown(context.Context) error
}

type tcpKeepAliveListener struct {
	*net.TCPListener
	period time.Duration
}

func NewListener(o Options, ctx context.Context, lcfg net.ListenConfig) (net.Listener, error) {
	tc, err := NewTlsConfig(o.Tls)
	if err != nil {
		return nil, err
	}

	l, err := lcfg.Listen(ctx, "tcp", o.Address)
	if err != nil {
		return nil, err
	}

	if tc != nil {
		l = tls.NewListener(l, tc)
	}

	if !o.DisableTCPKeepAlives {
		period := o.TCPKeepAlivePeriod
		if period <= 0 {
			period = defaultTCPKeepAlivePeriod
		}

		l = tcpKeepAliveListener{
			TCPListener: l.(*net.TCPListener),
			period:      period,
		}
	}

	return l, nil
}

// NewServerChain produces the standard constructor chain for a server, primarily using configuration.
func NewServerChain(o Options, l log.Logger, pb ...xloghttp.ParameterBuilder) alice.Chain {
	chain := alice.New(
		ResponseHeaders{Header: o.Header}.Then,
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
