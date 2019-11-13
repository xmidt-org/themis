package xhttpserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultTCPKeepAlivePeriod time.Duration = 3 * time.Minute // the value used internally by net/http
)

// Releasable is implemented by connections returned by Listener that can be marked as freed without closing
// the connection.  Primarily, this is for hijacked connections that calling code no longer wants to count toward
// the Listener's max connections limit.
type Releasable interface {
	Release()
}

// TlsConn is the behavior of a TLS-specific net.Conn.  This interface is implemented by connections returned
// by Listener if a tls.Config is supplied in the options.
type TlsConn interface {
	CloseWrite() error
	ConnectionState() tls.ConnectionState
	Handshake() error
	OCSPResponse() []byte
	VerifyHostname(string) error
}

type tcpConn struct {
	*net.TCPConn

	releaseOnce sync.Once
	release     func()
}

func (tc tcpConn) Release() {
	if tc.release != nil {
		tc.releaseOnce.Do(tc.release)
	}
}

func (tc tcpConn) Close() error {
	if tc.release != nil {
		tc.releaseOnce.Do(tc.release)
	}

	return tc.TCPConn.Close()
}

type tlsConn struct {
	*tls.Conn

	releaseOnce sync.Once
	release     func()
}

func (tc tlsConn) Release() {
	if tc.release != nil {
		tc.releaseOnce.Do(tc.release)
	}
}

func (tc tlsConn) Close() error {
	if tc.release != nil {
		tc.releaseOnce.Do(tc.release)
	}

	return tc.Conn.Close()
}

// Listener is a configurable net.Listener that provides the following features via options:
//     A TCP keep-alive duration
//     A TLS configuration
//     A limit on the number of concurrent connections
//
// Only TCP connections are supported by this listener.
type Listener struct {
	tcpListener        *net.TCPListener
	tcpKeepAlivePeriod time.Duration
	tlsConfig          *tls.Config

	// maxConnections is the maximum number of active connections allowed through this listener.
	// if nonpositive, no limit is imposed.
	maxConnections int32

	// count is the current count of active connections.  unused if maxConnections is nonpositive.
	count int32
}

func (l *Listener) checkAccept() bool {
	if l.maxConnections > 0 {
		if atomic.AddInt32(&l.count, 1) > l.maxConnections {
			atomic.AddInt32(&l.count, -1)
			return false
		}
	}

	return true
}

func (l *Listener) releaseConn() {
	atomic.AddInt32(&l.count, -1)
}

func (l *Listener) configure(c *net.TCPConn) error {
	if l.tcpKeepAlivePeriod > 0 {
		err := c.SetKeepAlive(true)
		if err == nil {
			err = c.SetKeepAlivePeriod(l.tcpKeepAlivePeriod)
		}

		if err != nil {
			c.Close()
			return err
		}
	}

	return nil
}

func (l *Listener) Accept() (net.Conn, error) {
	for {
		conn, err := l.tcpListener.AcceptTCP()
		if err != nil {
			return nil, err
		}

		if !l.checkAccept() {
			conn.Close()
			continue
		}

		if err := l.configure(conn); err != nil {
			return nil, err
		}

		if l.maxConnections > 0 {
			if l.tlsConfig != nil {
				return tlsConn{
					Conn:    tls.Server(conn, l.tlsConfig),
					release: l.releaseConn,
				}, nil
			}

			return tcpConn{
				TCPConn: conn,
				release: l.releaseConn,
			}, nil
		} else if l.tlsConfig != nil {
			return tls.Server(conn, l.tlsConfig), nil
		} else {
			return conn, nil
		}
	}
}

func (l *Listener) Close() error {
	return l.tcpListener.Close()
}

func (l *Listener) Addr() net.Addr {
	return l.tcpListener.Addr()
}

// NewListener constructs a net.Listener appropriate for the server configuration.  This function
// binds to the address specified in the options or an autoselected address if that field is one
// of the values mentioned at https://godoc.org/net#Listen.
func NewListener(ctx context.Context, o Options, lcfg net.ListenConfig, tcfg *tls.Config) (*Listener, error) {
	network := o.Network
	if len(network) == 0 {
		network = "tcp"
	}

	l, err := lcfg.Listen(ctx, network, o.Address)
	if err != nil {
		return nil, err
	}

	tcpListener, ok := l.(*net.TCPListener)
	if !ok {
		l.Close()
		return nil, fmt.Errorf("Network [%s] and address [%s] does not result in a TCPListener", network, o.Address)
	}

	listener := &Listener{
		tcpListener:    tcpListener,
		tlsConfig:      tcfg,
		maxConnections: o.MaxConnections,
	}

	if !o.DisableTCPKeepAlives {
		period := o.TCPKeepAlivePeriod
		if period <= 0 {
			period = defaultTCPKeepAlivePeriod
		}

		listener.tcpKeepAlivePeriod = period
	}

	return listener, nil
}
