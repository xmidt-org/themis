// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package xhttpserver

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
	"strings"

	"go.uber.org/multierr"
)

var (
	ErrTlsCertificateRequired         = errors.New("Both a certificateFile and keyFile are required")
	ErrUnableToAddClientCACertificate = errors.New("Unable to add client CA certificate")
)

// PeerVerifyError represents a verification error for a particular certificate
type PeerVerifyError struct {
	Certificate *x509.Certificate
	Reason      string
}

func (pve PeerVerifyError) Error() string {
	return pve.Reason
}

// PeerVerifyOptions allows common checks against a client-side certificate to be configured externally.  Any constraint that matches
// will result in a valid peer cert.
type PeerVerifyOptions struct {
	// DNSSuffixes enumerates any DNS suffixes that are checked.  A DNSName field of at least (1) peer cert
	// must have one of these suffixes.  If this field is not supplied, no DNS suffix checking is performed.
	// Matching is case insensitive.
	//
	// If any DNS suffix matches, that is sufficient for the peer cert to be valid.  No further checking is done in that case.
	DNSSuffixes []string

	// CommonNames lists the subject common names that at least (1) peer cert must have.  If not supplied,
	// no checking is done on the common name.  Matching common names is case sensitive.
	//
	// If any common name matches, that is sufficient for the peer cert to be valid.  No further checking is done in that case.
	CommonNames []string
}

// PeerVerifier is a verification strategy for a peer (client) certificate.
type PeerVerifier interface {
	Verify(peerCert *x509.Certificate, verifiedChains [][]*x509.Certificate) error
}

type PeerVerifierFunc func(*x509.Certificate, [][]*x509.Certificate) error

func (pvf PeerVerifierFunc) Verify(peerCert *x509.Certificate, verifiedChains [][]*x509.Certificate) error {
	return pvf(peerCert, verifiedChains)
}

// ConfiguredPeerVerifier is a PeerVerifier strategy synthesized from a PeerVerifyOptions.  This type is the built-in
// PeerVerifier strategy for this package.
type ConfiguredPeerVerifier struct {
	dnsSuffixes []string
	commonNames []string
}

func (cpv *ConfiguredPeerVerifier) Verify(peerCert *x509.Certificate, _ [][]*x509.Certificate) error {
	for _, suffix := range cpv.dnsSuffixes {
		for _, dnsName := range peerCert.DNSNames {
			if strings.HasSuffix(strings.ToLower(dnsName), suffix) {
				return nil
			}
		}

		// Allow the common name to be suffixed by a DNS suffix
		if strings.HasSuffix(strings.ToLower(peerCert.Subject.CommonName), suffix) {
			return nil
		}
	}

	for _, commonName := range cpv.commonNames {
		if commonName == peerCert.Subject.CommonName {
			return nil
		}
	}

	return PeerVerifyError{
		Certificate: peerCert,
		Reason:      "No DNS name or common name matched",
	}
}

// NewConfiguredPeerVerifier returns a ConfiguredPeerVerifier from a set of options.  If the given options
// do not represent any constraints, i.e. if every field is unset, then this function returns nil.
func NewConfiguredPeerVerifier(pvo PeerVerifyOptions) *ConfiguredPeerVerifier {
	if len(pvo.DNSSuffixes) == 0 && len(pvo.CommonNames) == 0 {
		return nil
	}

	cpv := new(ConfiguredPeerVerifier)
	if len(pvo.DNSSuffixes) > 0 {
		cpv.dnsSuffixes = make([]string, len(pvo.DNSSuffixes))
		for i, suffix := range pvo.DNSSuffixes {
			cpv.dnsSuffixes[i] = strings.ToLower(suffix)
		}
	}

	if len(pvo.CommonNames) > 0 {
		cpv.commonNames = append(cpv.commonNames, pvo.CommonNames...)
	}

	return cpv
}

// PeerVerifiers is a sequence of verification strategies.  All of the verifiers must return nil errors for
// a given peer cert to be considered valid.
type PeerVerifiers []PeerVerifier

// Verify allows a PeerVerifiers to itself be used as a PeerVerifier
func (pvs PeerVerifiers) Verify(peerCert *x509.Certificate, verifiedChains [][]*x509.Certificate) error {
	for _, pv := range pvs {
		if err := pv.Verify(peerCert, verifiedChains); err != nil {
			return err
		}
	}

	return nil
}

// VerifyPeerCertificate may be used as the closure for crypto/tls.Config.VerifyPeerCertificate
//
// If any of the rawCerts passes verification, this method returns nil to indicate that the
// client has supplied a valid certificate.  If all rawCerts fail verification or if any certificates
// fail to parse, this method returns an error.
func (pvs PeerVerifiers) VerifyPeerCertificate(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) (err error) {
	if len(pvs) == 0 {
		return
	}

	for _, rawCert := range rawCerts {
		peerCert, parseErr := x509.ParseCertificate(rawCert)
		if parseErr != nil {
			// any parse error invalidates the entire sequence of certificates
			err = parseErr
			break
		}

		verifyErr := pvs.Verify(peerCert, verifiedChains)
		err = multierr.Append(err, verifyErr)

		if verifyErr == nil {
			// we found (1) cert that passes verification, so we're good
			err = nil
			break
		}
	}

	return
}

// NewPeerVerifiers constructs a chain of verification strategies merged from a set of options with an extra
// set of application-layer strategies.  The extra verifiers are run first.  This function will return an empty
// chain of verifiers if both (1) the options do not have any constraints, and (2) there are no extra verifiers.
func NewPeerVerifiers(pvo PeerVerifyOptions, extra ...PeerVerifier) PeerVerifiers {
	pvs := append(PeerVerifiers{}, extra...)

	if cpv := NewConfiguredPeerVerifier(pvo); cpv != nil {
		pvs = append(pvs, cpv)
	}

	return pvs
}

// Tls represents the set of configurable options for a serverside tls.Config associated with a server.
type Tls struct {
	CertificateFile         string
	KeyFile                 string
	ClientCACertificateFile string
	ServerName              string
	NextProtos              []string
	MinVersion              uint16
	MaxVersion              uint16
	PeerVerify              PeerVerifyOptions
}

// NewTlsConfig produces a *tls.Config from a set of configuration options.  If the supplied set of options
// is nil, this function returns nil with no error.
//
// If supplied, the PeerVerifier strategies will be executed as part of peer verification.  This allows application-layer
// logic to be injected.
func NewTlsConfig(t *Tls, extra ...PeerVerifier) (*tls.Config, error) {
	if t == nil {
		return nil, nil
	}

	if len(t.CertificateFile) == 0 || len(t.KeyFile) == 0 {
		return nil, ErrTlsCertificateRequired
	}

	var nextProtos []string
	if len(t.NextProtos) > 0 {
		nextProtos = append(nextProtos, t.NextProtos...)
	} else {
		// assume http/1.1 by default
		nextProtos = append(nextProtos, "http/1.1")
	}

	tc := &tls.Config{ // nolint: gosec
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

	if pvs := NewPeerVerifiers(t.PeerVerify, extra...); len(pvs) > 0 {
		tc.VerifyPeerCertificate = pvs.VerifyPeerCertificate
	}

	if cert, err := tls.LoadX509KeyPair(t.CertificateFile, t.KeyFile); err != nil {
		return nil, err
	} else {
		tc.Certificates = []tls.Certificate{cert}
	}

	if len(t.ClientCACertificateFile) > 0 {
		caCert, err := os.ReadFile(t.ClientCACertificateFile)
		if err != nil {
			return nil, err
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, ErrUnableToAddClientCACertificate
		}

		tc.ClientCAs = caCertPool
		tc.ClientAuth = tls.RequestClientCert
	}

	tc.BuildNameToCertificate() // nolint: staticcheck
	return tc, nil
}
