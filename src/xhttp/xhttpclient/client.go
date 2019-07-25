package xhttpclient

import (
	"crypto/tls"
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

// NewRoundTripper assembles an http.RoundTripper given a set of configuration options.
// The returned round tripper will be backed by an http.Transport, decorated with any
// constructors that were supplied.  If the Transport options is nil, a default http.Transport
// is used.
func NewRoundTripper(t *Transport, c ...Constructor) (http.RoundTripper, error) {
	var delegate *http.Transport
	if t != nil {
		tc, err := NewTlsConfig(t.Tls)
		if err != nil {
			return nil, err
		}

		delegate = &http.Transport{
			DisableKeepAlives:      t.DisableKeepAlives,
			DisableCompression:     t.DisableCompression,
			MaxIdleConns:           t.MaxIdleConns,
			MaxIdleConnsPerHost:    t.MaxIdleConnsPerHost,
			MaxConnsPerHost:        t.MaxConnsPerHost,
			MaxResponseHeaderBytes: t.MaxResponseHeaderBytes,
			TLSClientConfig:        tc,
		}

		err = xerror.Do(
			xerror.TryOptionalDuration(t.IdleConnTimeout, &delegate.IdleConnTimeout),
			xerror.TryOptionalDuration(t.ResponseHeaderTimeout, &delegate.ResponseHeaderTimeout),
			xerror.TryOptionalDuration(t.ExpectContinueTimeout, &delegate.ExpectContinueTimeout),
			xerror.TryOptionalDuration(t.TlsHandshakeTimeout, &delegate.TLSHandshakeTimeout),
		)

		if err != nil {
			return nil, err
		}
	}

	return NewChain(c...).Then(delegate), nil
}

// New assembles an http client from a set of configuration options
func New(o Options, c ...Constructor) (Interface, error) {
	roundTripper, err := NewRoundTripper(o.Transport, c...)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: roundTripper,
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
