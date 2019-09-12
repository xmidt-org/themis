package xhttpserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddressKey(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(addressKey, AddressKey())
}

func TestServerKey(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(serverKey, ServerKey())
}
