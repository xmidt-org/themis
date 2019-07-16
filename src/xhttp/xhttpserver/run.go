package xhttpserver

import (
	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

type RouterBuilder func(*mux.Router, log.Logger) error

// Run is meant to be used in an fx.Invoke option.  This function uses Unmarshal to setup
// an HTTP server from configuration, then uses zero or more builders to configure the *mux.Router
// for that server.
func Run(configKey string, in ServerIn, b ...RouterBuilder) error {
	router, logger, err := Unmarshal(configKey, in)
	if err != nil {
		return err
	}

	for _, f := range b {
		if err := f(router, logger); err != nil {
			return err
		}
	}

	return nil
}

// Optional accepts an error from Run and returns nil if that error indicates the server is
// simply not configured, or the original error otherwise.  Wrapping a call to Run in this method
// allows a server to be optionally available via configuration.
func Optional(err error) error {
	if _, ok := err.(ServerNotConfiguredError); ok {
		return nil
	}

	return err
}
