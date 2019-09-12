package xhttpserver

import "github.com/go-kit/kit/log"

const (
	addressKey = "address"
	serverKey  = "server"
)

// AddressKey is the logging key for the server's bind address
func AddressKey() interface{} {
	return addressKey
}

// ServerKey is the logging key for the server's name
func ServerKey() interface{} {
	return serverKey
}

// NewServerLogger returns a go-kit Logger enriched with information about the server.
func NewServerLogger(key string, o Options, base log.Logger, extra ...interface{}) log.Logger {
	var parameters []interface{}
	if len(key) > 0 {
		parameters = append(parameters, ServerKey(), key)
	}

	parameters = append(parameters, extra...)
	if len(parameters) > 0 {
		return log.WithPrefix(base, parameters...)
	}

	return base
}
