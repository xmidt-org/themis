package xhttpserver

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
)

var (
	ErrTlsCertificateRequired         = errors.New("Both a certificateFile and keyFile are required")
	ErrUnableToAddClientCACertificate = errors.New("Unable to add client CA certificate")
)

type InvalidCertificateError struct {
	Certificate *x509.Certificate
}

func (ice InvalidCertificateError) Error() string {
	return fmt.Sprintf(
		"Certificate with common name [%s] and DNS names {%s} is not valid",
		ice.Certificate.Subject.CommonName,
		strings.Join(ice.Certificate.DNSNames, ","),
	)
}

// PeerVerify allows common checks against a client-side certificate to be configured externally.  Any constraint that matches
// will result in a valid peer cert.
type PeerVerify struct {
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
type PeerVerifier func(peerCert *x509.Certificate, verifiedChains [][]*x509.Certificate) error

// peerVerifier is the internal implementation of crypto/tls.Config.VerifyPeerCertificate
type peerVerifier struct {
	dnsSuffixes []string
	commonNames []string
	extra       []PeerVerifier
}

func (pv *peerVerifier) verifyParsedCertificate(cert *x509.Certificate, verifiedChains [][]*x509.Certificate) error {
	// always give application-layer code power of veto first ...
	for _, ef := range pv.extra {
		if err := ef(cert, verifiedChains); err != nil {
			return err
		}
	}

	for _, suffix := range pv.dnsSuffixes {
		for _, dnsName := range cert.DNSNames {
			if strings.HasSuffix(strings.ToLower(dnsName), suffix) {
				return nil
			}
		}
	}

	for _, commonName := range pv.commonNames {
		if cert.Subject.CommonName == commonName {
			return nil
		}
	}

	return InvalidCertificateError{Certificate: cert}
}

func (pv *peerVerifier) verifyPeerCertificate(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	// once we've verified *any* cert, we don't continue to verify.  we do, however, want to keep
	// parsing certs since if any cert isn't valid, we want to fail the verification.
	verified := false

	for _, rawCert := range rawCerts {
		cert, err := x509.ParseCertificate(rawCert)
		if err != nil {
			return err
		}

		if !verified {
			err = pv.verifyParsedCertificate(cert, verifiedChains)
			if err != nil {
				return err
			}

			verified = true
		}
	}

	return nil
}

// NewVerifyPeerCertificate produces a peer cert verification closure given configuration and application-layer logic.
// This function will return a nil function if no peer verification is configured in the Tls instance AND there are no
// application-layer PeerVerifiers supplied.
func NewVerifyPeerCertificate(t *Tls, extra ...PeerVerifier) func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	pv := &peerVerifier{
		extra: extra,
	}

	if t != nil {
		if len(t.PeerVerify.DNSSuffixes) > 0 {
			pv.dnsSuffixes = append(pv.dnsSuffixes, t.PeerVerify.DNSSuffixes...)
			for i := range pv.dnsSuffixes {
				pv.dnsSuffixes[i] = strings.ToLower(pv.dnsSuffixes[i])
			}
		}

		if len(t.PeerVerify.CommonNames) > 0 {
			pv.commonNames = append(pv.commonNames, t.PeerVerify.CommonNames...)
		}
	}

	if len(pv.extra) == 0 && len(pv.dnsSuffixes) == 0 && len(pv.commonNames) == 0 {
		return nil
	}

	return pv.verifyPeerCertificate
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
	PeerVerify              PeerVerify
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
		for _, np := range t.NextProtos {
			nextProtos = append(nextProtos, np)
		}
	} else {
		// assume http/1.1 by default
		nextProtos = append(nextProtos, "http/1.1")
	}

	tc := &tls.Config{
		MinVersion:            t.MinVersion,
		MaxVersion:            t.MaxVersion,
		ServerName:            t.ServerName,
		NextProtos:            nextProtos,
		VerifyPeerCertificate: NewVerifyPeerCertificate(t, extra...),
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
