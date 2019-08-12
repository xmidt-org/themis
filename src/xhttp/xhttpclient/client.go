package xhttpclient

import (
	"crypto/tls"
	"net/http"
	"time"
)

// Interface defines the behavior of an HTTP client.  *http.Client implements this interface.
type Interface interface {
	Do(*http.Request) (*http.Response, error)
}

// Tls represents the set of configurable options for client-side TLS
type Tls struct {
	InsecureSkipVerify bool
}

// Transport represents the set of configurable options for a client RoundTripper
// The majority of these fields map directory to an http.Transport.
// See https://godoc.org/net/http#Transport
type Transport struct {
	DisableKeepAlives      bool
	DisableCompression     bool
	MaxIdleConns           int
	MaxIdleConnsPerHost    int
	MaxConnsPerHost        int
	IdleConnTimeout        time.Duration
	ResponseHeaderTimeout  time.Duration
	ExpectContinueTimeout  time.Duration
	MaxResponseHeaderBytes int64
	TlsHandshakeTimeout    time.Duration
	Tls                    *Tls

	// Header is a set of static HTTP headers added to every request
	Header http.Header
}

// Options represents the set of configurable options for an HTTP client
type Options struct {
	Timeout   time.Duration
	Transport *Transport
}

// NewTlsConfig assembles a *tls.Config for clients given a set of configuration options.
// If the Tls options is nil, this method returns nil, nil.
func NewTlsConfig(tc *Tls) *tls.Config {
	if tc == nil {
		return nil
	}

	return &tls.Config{
		InsecureSkipVerify: tc.InsecureSkipVerify,
	}
}

// NewRoundTripper assembles an http.RoundTripper given a set of configuration options.
// The returned round tripper will be backed by an http.Transport, decorated with any
// constructors that were supplied.  If the Transport options is nil, a default http.Transport
// is used.
func NewRoundTripper(t *Transport, c ...Constructor) http.RoundTripper {
	var (
		delegate *http.Transport
		chain    Chain
	)

	if t != nil {
		chain = chain.Append(RequestHeaders{Header: t.Header}.Then)

		delegate = &http.Transport{
			DisableKeepAlives:      t.DisableKeepAlives,
			DisableCompression:     t.DisableCompression,
			MaxIdleConns:           t.MaxIdleConns,
			MaxIdleConnsPerHost:    t.MaxIdleConnsPerHost,
			MaxConnsPerHost:        t.MaxConnsPerHost,
			MaxResponseHeaderBytes: t.MaxResponseHeaderBytes,

			IdleConnTimeout:       t.IdleConnTimeout,
			ResponseHeaderTimeout: t.ResponseHeaderTimeout,
			ExpectContinueTimeout: t.ExpectContinueTimeout,
			TLSHandshakeTimeout:   t.TlsHandshakeTimeout,

			TLSClientConfig: NewTlsConfig(t.Tls),
		}
	} else {
		delegate = new(http.Transport)
	}

	return chain.Append(c...).Then(delegate)
}

// New assembles an http client from a set of configuration options
func New(o Options, c ...Constructor) Interface {
	return &http.Client{
		Transport: NewRoundTripper(o.Transport, c...),
		Timeout:   o.Timeout,
	}
}
