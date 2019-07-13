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
