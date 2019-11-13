package xhttpserver

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"testing"
)

// generateX509Cert returns an ASN.1 DER formatted x509 certificate together with the private key used to create it
func generateX509Cert(t *testing.T) (*x509.Certificate, *rsa.PrivateKey, []byte) {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Fatalf("Unable to generate private key: %s", err)
	}

	template := &x509.Certificate{
		Subject: pkix.Name{
			CommonName: "Comcast CA",
		},
		DNSNames: []string{"comcast.net"},
	}

	raw, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("Unable to create self-signed certificate: %s", err)
	}

	parsed, err := x509.ParseCertificate(raw)
	if err != nil {
		t.Fatalf("Unable to parse generated certificate: %s", err)
	}

	return parsed, key, raw
}

// encodeTlsCert uses generateX509Cert to create a certificate and private key appropriate for a tls.Config.
// The returned byte slices are PEM-coded.
func encodeTlsCert(t *testing.T) (pemCert, pemKey []byte) {
	_, pk, rawCert := generateX509Cert(t)

	rawKey, err := x509.MarshalPKCS8PrivateKey(pk)
	if err != nil {
		t.Fatalf("Unable to marshal private key: %s", err)
	}

	pemKey = pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: rawKey,
	})

	pemCert = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: rawCert,
	})

	return
}

// newGetCertificate produces a tls.Config.GetCertificate closure with a generated cert.  The returned
// cert will be self-signed.
func newGetCertificate(t *testing.T) func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	cert, err := tls.X509KeyPair(encodeTlsCert(t))
	if err != nil {
		t.Fatalf("Unable to create X509 key pair: %s", err)
	}

	return func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
		return &cert, err
	}
}
