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
	Header                 http.Header
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

// NewRoundTripper creates an http.RoundTripper from a set of Transport options.  If the Transport
// is nil, this function returns a default http.Transport instance.  Otherwise, an http.Transport
// is returned with its fields set from the given Transport options.
func NewRoundTripper(t *Transport) http.RoundTripper {
	var rt http.RoundTripper
	if t == nil {
		rt = new(http.Transport)
	} else {
		rt = RequestHeaders{Header: t.Header}.Then(&http.Transport{
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
		})
	}

	return rt
}

// New fully constructs an http client from a set of options.  NewRoundTripper is used to create the http.RoundTripper.
func New(o Options) Interface {
	return NewCustom(o, NewRoundTripper(o.Transport))
}

// NewCustom uses a set of options and a supplied RoundTripper to create an http client.  Use this function
// when a custom RoundTripper, including decoration, is desired.
func NewCustom(o Options, rt http.RoundTripper) Interface {
	return &http.Client{
		Transport: rt,
		Timeout:   o.Timeout,
	}
}
