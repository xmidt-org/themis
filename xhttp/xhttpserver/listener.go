package xhttpserver

import (
	"context"
	"crypto/tls"
	"net"
	"time"
)

const (
	defaultTCPKeepAlivePeriod time.Duration = 3 * time.Minute // the value used internally by net/http
)

// tcpKeepAliveListener mimics the internal type of the same name in net/http.
// This listener is configurable, and the period ultimately comes from server Options.
type tcpKeepAliveListener struct {
	*net.TCPListener
	period time.Duration
}

func (l tcpKeepAliveListener) Accept() (net.Conn, error) {
	return l.AcceptTCP()
}

func (l tcpKeepAliveListener) AcceptTCP() (*net.TCPConn, error) {
	conn, err := l.TCPListener.AcceptTCP()
	if err != nil {
		return nil, err
	}

	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(l.period)
	return conn, nil
}

// NewListener constructs a net.Listener appropriate for the server configuration.  This function
// binds to the address specified in the options or an autoselected address if that field is one
// of the values mentioned at https://godoc.org/net#Listen.
func NewListener(ctx context.Context, o Options, lcfg net.ListenConfig, extra ...PeerVerifier) (net.Listener, error) {
	tc, err := NewTlsConfig(o.Tls, extra...)
	if err != nil {
		return nil, err
	}

	network := o.Network
	if len(network) == 0 {
		network = "tcp"
	}

	l, err := lcfg.Listen(ctx, network, o.Address)
	if err != nil {
		return nil, err
	}

	// decorate with the keep alive first, since it needs a concrete type of listener
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

	if tc != nil {
		l = tls.NewListener(l, tc)
	}

	return l, nil
}
