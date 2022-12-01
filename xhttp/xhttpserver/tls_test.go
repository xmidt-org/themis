package xhttpserver

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestPeerVerifyError(t *testing.T) {
	var (
		assert = assert.New(t)
		err    = PeerVerifyError{Reason: "expected"}
	)

	assert.Equal("expected", err.Error())
}

func TestPeerVerifierFunc(t *testing.T) {
	testCert := new(x509.Certificate)
	pvf := PeerVerifierFunc(func(actual *x509.Certificate, verifiedChains [][]*x509.Certificate) error {
		assert.Equal(t, testCert, actual)
		return nil
	})

	assert.NoError(t, pvf.Verify(testCert, nil))
}

type ConfiguredPeerVerifierSuite struct {
	suite.Suite
}

func (suite *ConfiguredPeerVerifierSuite) newConfiguredPeerVerifier(o PeerVerifyOptions) *ConfiguredPeerVerifier {
	cpv := NewConfiguredPeerVerifier(o)
	suite.Require().NotNil(cpv)
	return cpv
}

func (suite *ConfiguredPeerVerifierSuite) testVerifySuccess() {
	testData := []struct {
		description string
		peerCert    x509.Certificate
		options     PeerVerifyOptions
	}{
		{
			description: "DNS name",
			peerCert: x509.Certificate{
				DNSNames: []string{"test.foobar.com"},
			},
			options: PeerVerifyOptions{
				DNSSuffixes: []string{"foobar.com"},
			},
		},
		{
			description: "multiple DNS names",
			peerCert: x509.Certificate{
				DNSNames: []string{"first.foobar.com", "second.something.net"},
			},
			options: PeerVerifyOptions{
				DNSSuffixes: []string{"another.thing.org", "something.net"},
			},
		},
		{
			description: "common name as host name",
			peerCert: x509.Certificate{
				Subject: pkix.Name{
					CommonName: "PCTEST-another.thing.org",
				},
			},
			options: PeerVerifyOptions{
				DNSSuffixes: []string{"another.thing.org", "something.net"},
			},
		},
		{
			description: "common name",
			peerCert: x509.Certificate{
				Subject: pkix.Name{
					CommonName: "A Great Organization",
				},
			},
			options: PeerVerifyOptions{
				CommonNames: []string{"A Great Organization"},
			},
		},
		{
			description: "multiple common names",
			peerCert: x509.Certificate{
				Subject: pkix.Name{
					CommonName: "A Great Organization",
				},
			},
			options: PeerVerifyOptions{
				CommonNames: []string{"First Organization Doesn't Match", "A Great Organization"},
			},
		},
	}

	for _, testCase := range testData {
		suite.Run(testCase.description, func() {
			verifier := suite.newConfiguredPeerVerifier(testCase.options)
			suite.NoError(verifier.Verify(&testCase.peerCert, nil))
		})
	}
}

func (suite *ConfiguredPeerVerifierSuite) testVerifyFailure() {
	testData := []struct {
		description string
		peerCert    x509.Certificate
		options     PeerVerifyOptions
	}{
		{
			description: "empty fields",
			peerCert:    x509.Certificate{},
			options: PeerVerifyOptions{
				DNSSuffixes: []string{"foobar.net"},
				CommonNames: []string{"For Great Justice"},
			},
		},
		{
			description: "DNS mismatch",
			peerCert: x509.Certificate{
				DNSNames: []string{"another.company.com"},
			},
			options: PeerVerifyOptions{
				DNSSuffixes: []string{"foobar.net"},
				CommonNames: []string{"For Great Justice"},
			},
		},
		{
			description: "CommonName mismatch",
			peerCert: x509.Certificate{
				Subject: pkix.Name{
					CommonName: "Villains For Hire",
				},
			},
			options: PeerVerifyOptions{
				DNSSuffixes: []string{"foobar.net"},
				CommonNames: []string{"For Great Justice"},
			},
		},
		{
			description: "DNS and CommonName mismatch",
			peerCert: x509.Certificate{
				DNSNames: []string{"another.company.com"},
				Subject: pkix.Name{
					CommonName: "Villains For Hire",
				},
			},
			options: PeerVerifyOptions{
				DNSSuffixes: []string{"foobar.net"},
				CommonNames: []string{"For Great Justice"},
			},
		},
	}

	for _, testCase := range testData {
		suite.Run(testCase.description, func() {
			verifier := suite.newConfiguredPeerVerifier(testCase.options)

			err := verifier.Verify(&testCase.peerCert, nil)
			suite.Error(err)
		})
	}
}

func (suite *ConfiguredPeerVerifierSuite) TestVerify() {
	suite.Run("Success", suite.testVerifySuccess)
	suite.Run("Failure", suite.testVerifyFailure)
}

func (suite *ConfiguredPeerVerifierSuite) TestNewConfiguredPeerVerifier() {
	suite.Run("Nil", func() {
		suite.Nil(NewConfiguredPeerVerifier(PeerVerifyOptions{}))
	})
}

func TestConfiguredPeerVerifier(t *testing.T) {
	suite.Run(t, new(ConfiguredPeerVerifierSuite))
}

type PeerVerifiersSuite struct {
	suite.Suite

	key       *rsa.PrivateKey
	testCerts []*x509.Certificate
	rawCerts  [][]byte
}

func (suite *PeerVerifiersSuite) SetupSuite() {
	var err error
	suite.key, err = rsa.GenerateKey(rand.Reader, 512) // nolint: gosec
	suite.Require().NoError(err)

	suite.testCerts = make([]*x509.Certificate, 3)
	suite.rawCerts = make([][]byte, len(suite.testCerts))

	for i := 0; i < len(suite.testCerts); i++ {
		suite.testCerts[i] = &x509.Certificate{
			SerialNumber: big.NewInt(int64(i + 1)),
			DNSNames: []string{
				fmt.Sprintf("host-%d.net", i),
			},
			Subject: pkix.Name{
				CommonName: fmt.Sprintf("Organization #%d", i),
			},
		}

		var err error
		suite.rawCerts[i], err = x509.CreateCertificate(
			rand.Reader,
			suite.testCerts[i],
			suite.testCerts[i],
			&suite.key.PublicKey,
			suite.key,
		)

		suite.Require().NoError(err)
		suite.Require().NotEmpty(suite.rawCerts[i])
	}
}

// useCertificate gives some syntactic sugar for expecting a peer cert
func (suite *PeerVerifiersSuite) useCertificate(expected *x509.Certificate) func(*x509.Certificate) bool {
	return newCertificateMatcher(
		suite.T(),
		expected.Subject.CommonName,
		expected.DNSNames...,
	)
}

func (suite *PeerVerifiersSuite) expectVerify(expected *x509.Certificate, result error) *mockPeerVerifier {
	m := new(mockPeerVerifier)
	m.ExpectVerify(
		suite.useCertificate(expected),
	).Return(result)

	return m
}

func (suite *PeerVerifiersSuite) TestUnparseableCertificate() {
	var (
		unparseable = []byte("unparseable")

		m  = new(mockPeerVerifier) // no calls
		pv = PeerVerifiers{m}
	)

	suite.Error(pv.VerifyPeerCertificate([][]byte{unparseable}, nil))
	m.AssertExpectations(suite.T())
}

func (suite *PeerVerifiersSuite) testVerifySuccess(expected *x509.Certificate) {
	for count := 0; count < 3; count++ {
		suite.Run(fmt.Sprintf("verifiers=%d", count), func() {
			var pv PeerVerifiers
			for i := 0; i < count; i++ {
				pv = append(pv, suite.expectVerify(expected, nil))
			}

			suite.NoError(
				pv.Verify(expected, nil),
			)

			assertPeerVerifierExpectations(suite.T(), pv...)
		})
	}
}

func (suite *PeerVerifiersSuite) testVerifyFailure(expected *x509.Certificate) {
	for count := 0; count < 3; count++ {
		suite.Run(fmt.Sprintf("goodVerifiers=%d", count), func() {
			var pv PeerVerifiers

			// setup our "good" calls
			for i := 0; i < count; i++ {
				pv = append(pv, suite.expectVerify(expected, nil))
			}

			// a failure, followed by a verifier that shouldn't be called
			pv = append(pv,
				suite.expectVerify(expected, errors.New("expected")),
				new(mockPeerVerifier),
			)

			suite.Error(
				pv.Verify(expected, nil),
			)

			assertPeerVerifierExpectations(suite.T(), pv...)
		})
	}
}

func (suite *PeerVerifiersSuite) TestVerify() {
	suite.Run("Success", func() {
		suite.testVerifySuccess(suite.testCerts[0])
	})

	suite.Run("Failure", func() {
		suite.testVerifyFailure(suite.testCerts[0])
	})
}

func (suite *PeerVerifiersSuite) testVerifyPeerCertificateAllGood(testCerts []*x509.Certificate, rawCerts [][]byte) {
	for count := 0; count < 3; count++ {
		suite.Run(fmt.Sprintf("verifiers=%d", count), func() {
			var pv PeerVerifiers
			for i := 0; i < count; i++ {
				m := new(mockPeerVerifier)
				for j := 0; j < len(testCerts); j++ {
					// maybe is used here because a success short-circuits subsequent calls
					m.ExpectVerify(suite.useCertificate(testCerts[j])).Return(error(nil)).Maybe()
				}

				pv = append(pv, m)
			}

			suite.NoError(
				pv.VerifyPeerCertificate(rawCerts, nil),
			)

			assertPeerVerifierExpectations(suite.T(), pv...)
		})
	}
}

func (suite *PeerVerifiersSuite) testVerifyPeerCertificateAllBad(testCerts []*x509.Certificate, rawCerts [][]byte) {
	for count := 1; count < 3; count++ {
		suite.Run(fmt.Sprintf("verifiers=%d", count), func() {
			var pv PeerVerifiers
			for i := 0; i < count; i++ {
				m := new(mockPeerVerifier)
				for j := 0; j < len(testCerts); j++ {
					// maybe is used here because a failure short-circuits subsequent calls
					m.ExpectVerify(suite.useCertificate(testCerts[j])).Return(errors.New("expected")).Maybe()
				}

				pv = append(pv, m)
			}

			suite.Error(
				pv.VerifyPeerCertificate(rawCerts, nil),
			)

			assertPeerVerifierExpectations(suite.T(), pv...)
		})
	}
}

func (suite *PeerVerifiersSuite) testVerifyPeerCertificateOneGood(testCerts []*x509.Certificate, rawCerts [][]byte) {
	// a verifier that passes any but the first cert
	oneGood := new(mockPeerVerifier)
	oneGood.ExpectVerify(func(actual *x509.Certificate) bool {
		return actual == testCerts[0]
	}).Return(errors.New("oneGood: first cert should fail")).Maybe()
	oneGood.ExpectVerify(func(actual *x509.Certificate) bool {
		return actual != testCerts[0]
	}).Return(error(nil))

	pv := PeerVerifiers{
		oneGood,
	}

	suite.NoError(
		pv.VerifyPeerCertificate(rawCerts, nil),
	)

	assertPeerVerifierExpectations(suite.T(), pv...)
}

func (suite *PeerVerifiersSuite) TestVerifyPeerCertificate() {
	suite.Run("AllGood", func() {
		for i := 1; i < len(suite.testCerts); i++ {
			suite.Run(fmt.Sprintf("certs=%d", i), func() {
				suite.testVerifyPeerCertificateAllGood(suite.testCerts[0:i], suite.rawCerts[0:i])
			})
		}
	})

	suite.Run("AllBad", func() {
		for i := 1; i < len(suite.testCerts); i++ {
			suite.Run(fmt.Sprintf("certs=%d", i), func() {
				suite.testVerifyPeerCertificateAllBad(suite.testCerts[0:i], suite.rawCerts[0:i])
			})
		}
	})

	suite.Run("OneGood", func() {
		for i := 2; i < len(suite.testCerts); i++ {
			suite.Run(fmt.Sprintf("certs=%d", i), func() {
				suite.testVerifyPeerCertificateOneGood(suite.testCerts[0:i], suite.rawCerts[0:i])
			})
		}
	})
}

func TestPeerVerifiers(t *testing.T) {
	suite.Run(t, new(PeerVerifiersSuite))
}

func TestNewPeerVerifiers(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		var (
			assert = assert.New(t)
			pv     = NewPeerVerifiers(PeerVerifyOptions{})
		)

		assert.Len(pv, 0)
	})

	t.Run("Configured", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)
			pv      = NewPeerVerifiers(PeerVerifyOptions{DNSSuffixes: []string{"foobar.com"}})
		)

		require.Len(pv, 1)
		assert.IsType((*ConfiguredPeerVerifier)(nil), pv[0])
	})

	t.Run("ConfiguredWithExtra", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)
			extra   = make([]PeerVerifier, 2)

			pv = NewPeerVerifiers(
				PeerVerifyOptions{DNSSuffixes: []string{"foobar.com"}},
				extra...,
			)
		)

		require.Len(pv, 3)
		assert.IsType((*ConfiguredPeerVerifier)(nil), pv[2])
	})
}

func testNewTlsConfigNil(t *testing.T) {
	assert := assert.New(t)
	tc, err := NewTlsConfig(nil)
	assert.Nil(tc)
	assert.NoError(err)
}

func testNewTlsConfigNoCertificateFile(t *testing.T) {
	assert := assert.New(t)
	tc, err := NewTlsConfig(&Tls{KeyFile: "key.pem"})
	assert.Nil(tc)
	assert.Equal(ErrTlsCertificateRequired, err)
}

func testNewTlsConfigNoKeyFile(t *testing.T) {
	assert := assert.New(t)
	tc, err := NewTlsConfig(&Tls{CertificateFile: "cert.pem"})
	assert.Nil(tc)
	assert.Equal(ErrTlsCertificateRequired, err)
}

func testNewTlsConfigLoadCertificateError(t *testing.T) {
	assert := assert.New(t)
	tc, err := NewTlsConfig(&Tls{
		CertificateFile: "nosuch",
		KeyFile:         "nosuch",
	})

	assert.Nil(tc)
	assert.Error(err)
}

func testNewTlsConfigSimple(t *testing.T, certificateFile, keyFile string) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		tc, err = NewTlsConfig(&Tls{
			CertificateFile: certificateFile,
			KeyFile:         keyFile,
		})
	)

	require.NoError(err)
	require.NotNil(tc)

	assert.Equal(uint16(tls.VersionTLS12), tc.MinVersion)
	assert.Zero(tc.MaxVersion)
	assert.Empty(tc.ServerName)
	assert.Equal([]string{"http/1.1"}, tc.NextProtos)
	assert.NotEmpty(tc.Certificates)
}

func testNewTlsConfigWithoutClientCACertificateFile(t *testing.T, certificateFile, keyFile string) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		tc, err = NewTlsConfig(&Tls{
			MinVersion:      1,
			MaxVersion:      3,
			ServerName:      "test",
			CertificateFile: certificateFile,
			KeyFile:         keyFile,
			NextProtos:      []string{"http/1.0"},
		})
	)

	require.NoError(err)
	require.NotNil(tc)

	assert.Equal(uint16(1), tc.MinVersion)
	assert.Equal(uint16(3), tc.MaxVersion)
	assert.Equal("test", tc.ServerName)
	assert.Equal([]string{"http/1.0"}, tc.NextProtos)
	assert.NotEmpty(tc.Certificates)
}

func testNewTlsConfigWithClientCACertificateFile(t *testing.T, certificateFile, keyFile string) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		tc, err = NewTlsConfig(&Tls{
			CertificateFile:         certificateFile,
			KeyFile:                 keyFile,
			ClientCACertificateFile: certificateFile,
			PeerVerify: PeerVerifyOptions{
				CommonNames: []string{"Hippies, Inc."},
			},
		})
	)

	require.NoError(err)
	require.NotNil(tc)

	assert.Equal(uint16(tls.VersionTLS12), tc.MinVersion)
	assert.Zero(tc.MaxVersion)
	assert.Empty(tc.ServerName)
	assert.Equal([]string{"http/1.1"}, tc.NextProtos)
	assert.NotEmpty(tc.Certificates)
	assert.NotNil(tc.ClientCAs)
	assert.Equal(tls.RequireAndVerifyClientCert, tc.ClientAuth)
}

func testNewTlsConfigLoadClientCACertificateError(t *testing.T, certificateFile, keyFile string) {
	var (
		assert = assert.New(t)

		tc, err = NewTlsConfig(&Tls{
			CertificateFile:         certificateFile,
			KeyFile:                 keyFile,
			ClientCACertificateFile: "nosuch",
		})
	)

	assert.Nil(tc)
	assert.Error(err)
}

func testNewTlsConfigAppendClientCACertificateError(t *testing.T, certificateFile, keyFile string) {
	var (
		assert = assert.New(t)

		tc, err = NewTlsConfig(&Tls{
			CertificateFile:         certificateFile,
			KeyFile:                 keyFile,
			ClientCACertificateFile: keyFile, // not a certificate, but still valid PEM
		})
	)

	assert.Nil(tc)
	assert.Equal(ErrUnableToAddClientCACertificate, err)
}

func TestNewTlsConfig(t *testing.T) {
	certificateFile, keyFile := createServerFiles(t)
	defer os.Remove(certificateFile)
	defer os.Remove(keyFile)

	t.Logf("Using certificate file '%s' and key file '%s'", certificateFile, keyFile)

	t.Run("Nil", testNewTlsConfigNil)
	t.Run("NoCertificateFile", testNewTlsConfigNoCertificateFile)
	t.Run("NoKeyFile", testNewTlsConfigNoKeyFile)
	t.Run("LoadCertificateError", testNewTlsConfigLoadCertificateError)

	t.Run("Simple", func(t *testing.T) {
		testNewTlsConfigSimple(t, certificateFile, keyFile)
	})

	t.Run("WithoutClientCACertificateFile", func(t *testing.T) {
		testNewTlsConfigWithoutClientCACertificateFile(t, certificateFile, keyFile)
	})

	t.Run("WithClientCACertificateFile", func(t *testing.T) {
		testNewTlsConfigWithClientCACertificateFile(t, certificateFile, keyFile)
	})

	t.Run("LoadClientCACertificateError", func(t *testing.T) {
		testNewTlsConfigLoadClientCACertificateError(t, certificateFile, keyFile)
	})

	t.Run("AppendClientCACertificateError", func(t *testing.T) {
		testNewTlsConfigAppendClientCACertificateError(t, certificateFile, keyFile)
	})
}
