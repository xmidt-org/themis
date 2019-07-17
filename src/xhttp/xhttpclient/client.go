package xhttpclient

import (
	"crypto/tls"
	"log"
	"net/http"
	"time"
	"xerror"
)

// Interface defines the behavior of an HTTP client.  *http.Client implements this interface.
type Interface interface {
	Do(*http.Request) (*http.Response, error)
}

type Tls struct {
	InsecureSkipVerify bool
}

type Transport struct {
	DisableKeepAlives      bool
	DisableCompression     bool
	MaxIdleConns           int
	MaxIdleConnsPerHost    int
	MaxConnsPerHost        int
	IdleConnTimeout        string
	ResponseHeaderTimeout  string
	ExpectContinueTimeout  string
	MaxResponseHeaderBytes int64
	TlsHandshakeTimeout    string
	Tls                    *Tls
}

type Options struct {
	Timeout   string
	Transport *Transport
}

// NewTlsConfig assembles a *tls.Config for clients given a set of configuration options.
// If the Tls options is nil, this method returns nil, nil.
func NewTlsConfig(tc *Tls) (*tls.Config, error) {
	if tc == nil {
		return nil, nil
	}

	return &tls.Config{
		InsecureSkipVerify: tc.InsecureSkipVerify,
	}, nil
}

// NewTransport assembles an http.Transport given a set of configuration options.
// If the supplied Transport options is nil, this method returns a non-nil default transport
// that is distinct from net/http.
func NewTransport(t *Transport) (*http.Transport, error) {
	if t != nil {
		tc, err := NewTlsConfig(t.Tls)
		if err != nil {
			return nil, err
		}

		transport := &http.Transport{
			DisableKeepAlives:      t.DisableKeepAlives,
			DisableCompression:     t.DisableCompression,
			MaxIdleConns:           t.MaxIdleConns,
			MaxIdleConnsPerHost:    t.MaxIdleConnsPerHost,
			MaxConnsPerHost:        t.MaxConnsPerHost,
			MaxResponseHeaderBytes: t.MaxResponseHeaderBytes,
			TLSClientConfig:        tc,
		}

		err = xerror.Do(
			xerror.TryOptionalDuration(t.IdleConnTimeout, &transport.IdleConnTimeout),
			xerror.TryOptionalDuration(t.ResponseHeaderTimeout, &transport.ResponseHeaderTimeout),
			xerror.TryOptionalDuration(t.ExpectContinueTimeout, &transport.ExpectContinueTimeout),
			xerror.TryOptionalDuration(t.TlsHandshakeTimeout, &transport.TLSHandshakeTimeout),
		)

		if err != nil {
			return nil, err
		}

		return transport, nil
	}

	// always return a transport other than the default
	return new(http.Transport), nil
}

// New assembles an http client from a set of configuration options
func New(logger log.Logger, o Options) (Interface, error) {
	transport, err := NewTransport(o.Transport)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: transport,
	}

	if len(o.Timeout) > 0 {
		var err error
		client.Timeout, err = time.ParseDuration(o.Timeout)
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}
