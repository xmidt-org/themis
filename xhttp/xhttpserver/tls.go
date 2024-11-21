// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package xhttpserver

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
)

var (
	ErrTlsCertificateRequired = errors.New("Both a certificateFile and keyFile are required")
)

// ReadCertPool reads a file that is expected to contain a certificate bundle
// and returns that bundle as a pool.
func ReadCertPool(path string) (cp *x509.CertPool, err error) {
	var contents []byte
	contents, err = os.ReadFile(path)
	if err == nil {
		cp = x509.NewCertPool()
		if !cp.AppendCertsFromPEM(contents) {
			err = fmt.Errorf("Unable to add certificates from %s", path)
		}
	}

	return
}

// Mtls configures the mutual TLS settings for a tls.Config.
type Mtls struct {
	ClientCACertificateFile string
	DisableRequire          bool
	DisableVerify           bool
}

// Tls represents the set of configurable options for a serverside tls.Config associated with a server.
type Tls struct {
	CertificateFile string
	KeyFile         string
	Mtls            *Mtls
	ServerName      string
	NextProtos      []string
	MinVersion      uint16
	MaxVersion      uint16
}

// configureMtls sets up mtls on the given TLS configuration.
func configureMtls(tc *tls.Config, mtls *Mtls) (err error) {
	if mtls == nil {
		return
	}

	switch {
	case mtls.DisableRequire && mtls.DisableVerify:
		tc.ClientAuth = tls.RequestClientCert

	case !mtls.DisableRequire && mtls.DisableVerify:
		tc.ClientAuth = tls.RequireAnyClientCert

	case mtls.DisableRequire && !mtls.DisableVerify:
		tc.ClientAuth = tls.VerifyClientCertIfGiven

	case !mtls.DisableRequire && !mtls.DisableVerify:
		tc.ClientAuth = tls.RequireAndVerifyClientCert
	}

	tc.ClientCAs, err = ReadCertPool(mtls.ClientCACertificateFile)
	return
}

// NewTlsConfig produces a *tls.Config from a set of configuration options.  If the supplied set of options
// is nil, this function returns nil with no error.
func NewTlsConfig(t *Tls) (tc *tls.Config, err error) {
	if t == nil {
		return
	} else if len(t.CertificateFile) == 0 || len(t.KeyFile) == 0 {
		err = ErrTlsCertificateRequired
	}

	if err == nil {
		var nextProtos []string
		if len(t.NextProtos) > 0 {
			nextProtos = append(nextProtos, t.NextProtos...)
		} else {
			// assume http/1.1 by default
			nextProtos = append(nextProtos, "http/1.1")
		}

		tc = &tls.Config{ // nolint: gosec
			MinVersion: t.MinVersion,
			MaxVersion: t.MaxVersion,
			ServerName: t.ServerName,
			NextProtos: nextProtos,

			// disable vulnerable ciphers
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			},
		}

		// if no MinVersion was set, default to TLS 1.2
		if tc.MinVersion == 0 {
			tc.MinVersion = tls.VersionTLS12
		}

		tc.Certificates = make([]tls.Certificate, 1)
		tc.Certificates[0], err = tls.LoadX509KeyPair(t.CertificateFile, t.KeyFile)
	}

	if err == nil {
		err = configureMtls(tc, t.Mtls)
	}

	return
}
