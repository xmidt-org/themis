// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package xhttpserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
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

// Listener is a configurable net.Listener that provides the following features via options
type Listener struct {
	tcpListener        *net.TCPListener
	tcpKeepAlivePeriod time.Duration
	tlsConfig          *tls.Config
}

func (l *Listener) Accept() (net.Conn, error) {
	conn, err := l.tcpListener.AcceptTCP()
	if err != nil {
		return nil, err
	}

	if l.tcpKeepAlivePeriod > 0 {
		err := conn.SetKeepAlive(true)
		if err == nil {
			err = conn.SetKeepAlivePeriod(l.tcpKeepAlivePeriod)
		}

		if err != nil {
			conn.Close()
			return nil, err
		}
	}

	if l.tlsConfig != nil {
		return tls.Server(conn, l.tlsConfig), nil
	}

	return conn, nil
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
		tcpListener: tcpListener,
		tlsConfig:   tcfg,
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
