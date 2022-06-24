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

	key *rsa.PrivateKey
}

func (suite *PeerVerifiersSuite) SetupSuite() {
	var err error
	suite.key, err = rsa.GenerateKey(rand.Reader, 512)
	suite.Require().NoError(err)
}

func (suite *PeerVerifiersSuite) createSelfSignedCertificate(template *x509.Certificate) []byte {
	if template.SerialNumber == nil {
		template.SerialNumber = big.NewInt(1)
	}

	raw, err := x509.CreateCertificate(
		rand.Reader,
		template,
		template,
		suite.key.PublicKey,
		suite.key,
	)

	suite.Require().NoError(err)
	suite.Require().NotEmpty(raw)
	return raw
}

func (suite *PeerVerifiersSuite) TestUnparseableCertificate() {
	var (
		unparseable = []byte("unparseable")

		m = PeerVerifierFunc(func(*x509.Certificate, [][]*x509.Certificate) error {
			suite.Fail("This verifier should not have been called due to an unparseable certificate")
			return nil
		})

		pv = PeerVerifiers{m}
	)

	suite.Error(pv.VerifyPeerCertificate([][]byte{unparseable}, nil))
}

func (suite *PeerVerifiersSuite) testVerifySuccess() {
	for l := 0; l < 3; l++ {
		suite.Run(fmt.Sprintf("len=%d", l), func() {
			var (
				callCount int
				verifier  = PeerVerifierFunc(
					func(cert *x509.Certificate, _ [][]*x509.Certificate) error {
						callCount++
						return nil
					},
				)

				pv PeerVerifiers
			)

			for i := 0; i < l; i++ {
				pv = append(pv, verifier)
			}

			suite.NoError(pv.Verify(
				&x509.Certificate{},
				nil,
			))

			suite.Equal(l, callCount)
		})
	}
}

func (suite *PeerVerifiersSuite) testVerifyFailure() {
	for l := 1; l < 4; l++ {
		suite.Run(fmt.Sprintf("len=%d", l), func() {
			var (
				goodCount int
				good      = PeerVerifierFunc(
					func(cert *x509.Certificate, _ [][]*x509.Certificate) error {
						goodCount++
						return nil
					},
				)

				badCount int
				bad      = PeerVerifierFunc(
					func(cert *x509.Certificate, _ [][]*x509.Certificate) error {
						badCount++
						return errors.New("expected")
					},
				)

				shouldNotBeCalled = PeerVerifierFunc(
					func(cert *x509.Certificate, _ [][]*x509.Certificate) error {
						suite.Fail("This peer verifier should not have been called")
						return errors.New("this should not have been called")
					},
				)

				pv PeerVerifiers
			)

			for i := 0; i < l-1; i++ {
				pv = append(pv, good)
			}

			pv = append(pv, bad, shouldNotBeCalled)

			suite.Error(pv.Verify(
				&x509.Certificate{},
				nil,
			))

			suite.Equal(l-1, goodCount)
			suite.Equal(1, badCount)
		})
	}
}

func (suite *PeerVerifiersSuite) TestVerify() {
	suite.Run("Success", suite.testVerifySuccess)
	suite.Run("Failure", suite.testVerifyFailure)
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

	assert.Zero(tc.MinVersion)
	assert.Zero(tc.MaxVersion)
	assert.Empty(tc.ServerName)
	assert.Equal([]string{"http/1.1"}, tc.NextProtos)
	assert.NotEmpty(tc.NameToCertificate)
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
	assert.NotEmpty(tc.NameToCertificate)
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

	assert.Zero(tc.MinVersion)
	assert.Zero(tc.MaxVersion)
	assert.Empty(tc.ServerName)
	assert.Equal([]string{"http/1.1"}, tc.NextProtos)
	assert.NotEmpty(tc.NameToCertificate)
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
