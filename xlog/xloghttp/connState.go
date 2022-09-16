package xloghttp

import (
	"net"
	"net/http"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

// NewConnStateLogger produces an http/Server.ConnState function that logs the connection
// state to the supplied logger.
func NewConnStateLogger(logger log.Logger, key string, lvl level.Value) func(net.Conn, http.ConnState) {
	if lvl != nil {
		return func(_ net.Conn, cs http.ConnState) {
			logger.Log(level.Key(), lvl, key, cs.String())
		}
	}

	return func(_ net.Conn, cs http.ConnState) {
		logger.Log(key, cs.String())
	}
}
