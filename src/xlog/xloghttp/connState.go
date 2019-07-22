package xloghttp

import (
	"net"
	"net/http"

	"github.com/go-kit/kit/log"
)

// NewConnStateLogger produces an http/Server.ConnState function that logs the connection
// state to the supplied logger.
func NewConnStateLogger(logger log.Logger) func(net.Conn, http.ConnState) {
	return func(c net.Conn, cs http.ConnState) {
		logger.Log("state", cs.String())
	}
}
