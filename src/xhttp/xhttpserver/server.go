package xhttpserver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"time"
	"xlog"
	"xlog/xloghttp"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/justinas/alice"
)

const (
	addressKey = "address"
	serverKey  = "server"

	defaultTCPKeepAlivePeriod time.Duration = 3 * time.Minute // the value used internally by net/http
)

var (
	ErrNoAddress                      = errors.New("A server bind address must be specified")
	ErrTlsCertificateRequired         = errors.New("Both a certificateFile and keyFile are required")
	ErrUnableToAddClientCACertificate = errors.New("Unable to add client CA certificate")
)

// AddressKey is the logging key for the server's bind address
func AddressKey() interface{} {
	return addressKey
}

// ServerKey is the logging key for the server's name
func ServerKey() interface{} {
	return serverKey
}

type Tls struct {
	CertificateFile         string
	KeyFile                 string
	ClientCACertificateFile string
	ServerName              string
	NextProtos              []string
	MinVersion              uint16
	MaxVersion              uint16
}

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
	TCPKeepAlivePeriod   string

	Header http.Header
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

func NewTlsConfig(t *Tls) (*tls.Config, error) {
	if t == nil {
		return nil, nil
	}

	if len(t.CertificateFile) == 0 || len(t.KeyFile) == 0 {
		return nil, ErrTlsCertificateRequired
	}

	var nextProtos []string
	if len(t.NextProtos) > 0 {
		for _, np := range t.NextProtos {
			nextProtos = append(nextProtos, np)
		}
	} else {
		// assume http/1.1 by default
		nextProtos = append(nextProtos, "http/1.1")
	}

	tc := &tls.Config{
		MinVersion: t.MinVersion,
		MaxVersion: t.MaxVersion,
		ServerName: t.ServerName,
		NextProtos: nextProtos,
	}

	if cert, err := tls.LoadX509KeyPair(t.CertificateFile, t.KeyFile); err != nil {
		return nil, err
	} else {
		tc.Certificates = []tls.Certificate{cert}
	}

	if len(t.ClientCACertificateFile) > 0 {
		caCert, err := ioutil.ReadFile(t.ClientCACertificateFile)
		if err != nil {
			return nil, err
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, ErrUnableToAddClientCACertificate
		}

		tc.ClientCAs = caCertPool
		tc.ClientAuth = tls.RequireAndVerifyClientCert
	}

	tc.BuildNameToCertificate()
	return tc, nil
}

func NewListener(ctx context.Context, lcfg net.ListenConfig, o Options) (net.Listener, error) {
	address := o.Address
	if len(address) == 0 {
		address = ":http"
	}

	tc, err := NewTlsConfig(o.Tls)
	if err != nil {
		return nil, err
	}

	l, err := lcfg.Listen(ctx, "tcp", address)
	if err != nil {
		return nil, err
	}

	if tc != nil {
		l = tls.NewListener(l, tc)
	}

	if !o.DisableTCPKeepAlives {
		period := defaultTCPKeepAlivePeriod
		if len(o.TCPKeepAlivePeriod) > 0 {
			var err error
			period, err = time.ParseDuration(o.TCPKeepAlivePeriod)
			if err != nil {
				return nil, err
			}
		}

		l = tcpKeepAliveListener{
			TCPListener: l.(*net.TCPListener),
			period:      period,
		}
	}

	return l, nil
}

// OnStart produces a closure that will start the given server appropriately
func OnStart(logger log.Logger, s Interface, onExit func(), o Options) func(context.Context) error {
	if len(o.Address) == 0 {
		o.Address = ":http"
	}

	return func(ctx context.Context) error {
		l, err := NewListener(ctx, net.ListenConfig{}, o)
		if err != nil {
			return err
		}

		go func() {
			if onExit != nil {
				defer onExit()
			}

			logger.Log(level.Key(), level.InfoValue(), xlog.MessageKey(), "starting server")
			err := s.Serve(l)

			logger.Log(
				level.Key(), level.ErrorValue(),
				xlog.MessageKey(), "listener exited",
				xlog.ErrorKey(), err,
			)
		}()

		return nil
	}
}

// OnStop produces a closure that will shutdown the server appropriately
func OnStop(logger log.Logger, s Interface) func(context.Context) error {
	return func(ctx context.Context) error {
		logger.Log(
			level.Key(), level.InfoValue(),
			xlog.MessageKey(), "server stopping",
		)

		return s.Shutdown(ctx)
	}
}

// New constructs a basic HTTP server instance.  The supplied logger is enriched with information
// about the server and returned for use by higher-level code.
func New(base log.Logger, h http.Handler, o Options) (Interface, log.Logger) {
	if len(o.Address) == 0 {
		o.Address = ":http"
	}

	chain := alice.New(
		ResponseHeaders{Header: o.Header}.Then,
	)

	s := &http.Server{
		// we don't need this technically, because we create a listener
		// it's here for other code to inspect
		Addr:    o.Address,
		Handler: chain.Then(h),

		MaxHeaderBytes:    o.MaxHeaderBytes,
		IdleTimeout:       o.IdleTimeout,
		ReadHeaderTimeout: o.ReadHeaderTimeout,
		ReadTimeout:       o.ReadTimeout,
		WriteTimeout:      o.WriteTimeout,

		ErrorLog: xloghttp.NewErrorLog(
			o.Address,
			log.WithPrefix(
				base,
				level.Key(), level.ErrorValue(),
				AddressKey(), o.Address,
			),
		),
	}

	if o.LogConnectionState {
		s.ConnState = xloghttp.NewConnStateLogger(
			log.WithPrefix(
				base,
				AddressKey(), o.Address,
			),
			"connState",
			level.DebugValue(),
		)
	}

	if o.DisableHTTPKeepAlives {
		s.SetKeepAlivesEnabled(false)
	}

	return s, log.WithPrefix(base, ServerKey(), o.Name, AddressKey(), o.Address)
}
