// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package xhttpserver

import (
	"crypto/tls"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	_, err := NewTlsConfig(&Tls{
		CertificateFile: "nosuch",
		KeyFile:         "nosuch",
	})

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
			CertificateFile: certificateFile,
			KeyFile:         keyFile,
			Mtls: &Mtls{
				ClientCACertificateFile: certificateFile,
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

		_, err = NewTlsConfig(&Tls{
			CertificateFile: certificateFile,
			KeyFile:         keyFile,
			Mtls: &Mtls{
				ClientCACertificateFile: "nosuch",
			},
		})
	)

	assert.Error(err)
}

func testNewTlsConfigAppendClientCACertificateError(t *testing.T, certificateFile, keyFile string) {
	var (
		assert = assert.New(t)

		_, err = NewTlsConfig(&Tls{
			CertificateFile: certificateFile,
			KeyFile:         keyFile,
			Mtls: &Mtls{
				ClientCACertificateFile: keyFile, // not a certificate, but still valid PEM
			},
		})
	)

	assert.Error(err)
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
