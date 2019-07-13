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
	"xerror"
	"xlog"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

const (
	addressKey = "address"

	defaultTCPKeepAlivePeriod time.Duration = 3 * time.Minute // the value used internally by net/http
)

var (
	ErrNoAddress                      = errors.New("A server bind address must be specified")
	ErrTlsCertificateRequired         = errors.New("Both a certificateFile and keyFile are required")
	ErrUnableToAddClientCACertificate = errors.New("Unable to add client CA certificate")
)

func AddressKey() interface{} {
	return addressKey
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
	Address string
	Tls     *Tls

	LogConnectionState    bool
	DisableHTTPKeepAlives bool
	MaxHeaderBytes        int

	IdleTimeout       string
	ReadHeaderTimeout string
	ReadTimeout       string
	WriteTimeout      string

	DisableTCPKeepAlives bool
	TCPKeepAlivePeriod   string

	Header map[string]string
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
func New(base log.Logger, h http.Handler, o Options) (Interface, log.Logger, error) {
	if len(o.Address) == 0 {
		o.Address = ":http"
	}

	s := &http.Server{
		// we don't need this technically, because we create a listener
		// it's here for other code to inspect
		Addr: o.Address,

		Handler:        h,
		MaxHeaderBytes: o.MaxHeaderBytes,
		ErrorLog: xlog.NewErrorLog(
			o.Address,
			log.WithPrefix(
				base,
				level.Key(), level.ErrorValue(),
				AddressKey(), o.Address,
			),
		),
	}

	err := xerror.Do(
		func() (err error) {
			if len(o.IdleTimeout) > 0 {
				s.IdleTimeout, err = time.ParseDuration(o.IdleTimeout)
			}

			return
		},
		func() (err error) {
			if len(o.ReadHeaderTimeout) > 0 {
				s.ReadHeaderTimeout, err = time.ParseDuration(o.ReadHeaderTimeout)
			}

			return
		},
		func() (err error) {
			if len(o.ReadTimeout) > 0 {
				s.ReadTimeout, err = time.ParseDuration(o.ReadTimeout)
			}

			return
		},
		func() (err error) {
			if len(o.WriteTimeout) > 0 {
				s.WriteTimeout, err = time.ParseDuration(o.WriteTimeout)
			}

			return
		},
	)

	if err != nil {
		return nil, nil, err
	}

	if o.LogConnectionState {
		s.ConnState = xlog.NewConnStateLogger(
			log.WithPrefix(
				base,
				level.Key(), level.DebugValue(),
				AddressKey(), o.Address,
			),
		)
	}

	if o.DisableHTTPKeepAlives {
		s.SetKeepAlivesEnabled(false)
	}

	return s, log.WithPrefix(base, AddressKey(), o.Address), nil
}
