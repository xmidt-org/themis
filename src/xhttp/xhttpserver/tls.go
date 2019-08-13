package xhttpserver

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
)

var (
	ErrTlsCertificateRequired         = errors.New("Both a certificateFile and keyFile are required")
	ErrUnableToAddClientCACertificate = errors.New("Unable to add client CA certificate")
)

// Tls represents the set of configurable options for a serverside tls.Config associated with a server.
type Tls struct {
	CertificateFile         string
	KeyFile                 string
	ClientCACertificateFile string
	ServerName              string
	NextProtos              []string
	MinVersion              uint16
	MaxVersion              uint16
}

// NewTlsConfig produces a *tls.Config from a set of configuration options.  If the supplied set of options
// is nil, this function returns nil with no error.
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
