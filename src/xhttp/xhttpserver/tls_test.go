package xhttpserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestNewTlsConfig(t *testing.T) {
	t.Run("Nil", testNewTlsConfigNil)
	t.Run("NoCertificateFile", testNewTlsConfigNoCertificateFile)
	t.Run("NoKeyFile", testNewTlsConfigNoKeyFile)
	t.Run("LoadCertificateError", testNewTlsConfigLoadCertificateError)
}
