package xhttpserver

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"testing"
)

// generateTestCert returns an ASN.1 DER formatted x509 certificate together with the private key used to create it
func generateTestCert(t *testing.T) ([]byte, *rsa.PrivateKey) {
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

	return raw, key
}
