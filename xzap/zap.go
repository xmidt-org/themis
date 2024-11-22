// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package xzap

import (
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type stringArray []string

func (sa stringArray) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, s := range sa {
		enc.AppendString(s)
	}

	return nil
}

type pkixName pkix.Name

func (pn pkixName) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddArray("organization", stringArray(pn.Organization))
	enc.AddArray("organizationalUnit", stringArray(pn.OrganizationalUnit))
	enc.AddString("commonName", pn.CommonName)
	enc.AddString("serialNumber", pn.SerialNumber)

	return nil
}

type certificate x509.Certificate

func (c certificate) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddObject("issuer", pkixName(c.Issuer))
	enc.AddObject("subject", pkixName(c.Subject))
	enc.AddArray("dnsNames", stringArray(c.DNSNames))
	enc.AddArray("emailAddresses", stringArray(c.EmailAddresses))
	enc.AddArray("issuingCertificateURL", stringArray(c.IssuingCertificateURL))
	enc.AddTime("notBefore", c.NotBefore)
	enc.AddTime("notAfter", c.NotAfter)

	if c.SerialNumber != nil {
		enc.AddString("serialNumber", c.SerialNumber.String())
	} else {
		enc.AddString("serialNumber", "<none>")
	}

	enc.AddBinary("raw", c.Raw)
	return nil
}

func Certificate(field string, cert *x509.Certificate) zap.Field {
	if cert != nil {
		return zap.Object(field, certificate(*cert))
	} else {
		return zap.Skip()
	}
}

type certificates []*x509.Certificate

func (cs certificates) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, c := range cs {
		if c != nil {
			enc.AppendObject(certificate(*c))
		}
	}

	return nil
}

func tlsVersionToString(v uint16) string {
	switch v {
	case tls.VersionTLS10:
		return "1.0"

	case tls.VersionTLS11:
		return "1.1"

	case tls.VersionTLS12:
		return "1.2"

	case tls.VersionTLS13:
		return "1.3"

	case tls.VersionSSL30: //nolint:staticcheck
		return "SSLv3"

	default:
		return "unknown"
	}
}

type connectionState tls.ConnectionState

func (cstate connectionState) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("version", tlsVersionToString(cstate.Version))
	enc.AddArray("peerCertificates", certificates(cstate.PeerCertificates))

	return nil
}

// ConnectionState produces a zap logging Field that produces an object representation
// of a TLS connection state.
func ConnectionState(field string, v *tls.ConnectionState) zap.Field {
	if v != nil {
		return zap.Object(field, connectionState(*v))
	} else {
		return zap.Skip()
	}
}
