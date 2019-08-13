package xhttpserver

import (
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
	tc, err := NewTlsConfig(&Tls{
		CertificateFile: "nosuch",
		KeyFile:         "nosuch",
	})

	assert.Nil(tc)
	assert.Error(err)
}

func testNewTlsConfigSimple(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		tc, err = NewTlsConfig(&Tls{
			CertificateFile: "server.cert",
			KeyFile:         "server.key",
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

func testNewTlsConfigWithoutClientCACertificateFile(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		tc, err = NewTlsConfig(&Tls{
			MinVersion:      1,
			MaxVersion:      3,
			ServerName:      "test",
			CertificateFile: "server.cert",
			KeyFile:         "server.key",
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

func TestNewTlsConfig(t *testing.T) {
	t.Run("Nil", testNewTlsConfigNil)
	t.Run("NoCertificateFile", testNewTlsConfigNoCertificateFile)
	t.Run("NoKeyFile", testNewTlsConfigNoKeyFile)
	t.Run("LoadCertificateError", testNewTlsConfigLoadCertificateError)
	t.Run("Simple", testNewTlsConfigSimple)
	t.Run("WithoutClientCACertificateFile", testNewTlsConfigWithoutClientCACertificateFile)
}
