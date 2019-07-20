package xhttpserver

import (
	"xlog"

	"github.com/go-kit/kit/log/level"
)

// RunBuilder is a strategy for building up routes and other infrastructure related to an http.Server
type RunBuilder func(UnmarshalResult) error

// Run is meant to be used in an fx.Invoke option.  This function uses Unmarshal to setup
// an HTTP server from configuration, then uses zero or more builders to configure the *mux.Router
// for that server.
func Run(configKey string, in ServerIn, b ...RunBuilder) (UnmarshalResult, error) {
	result, err := Unmarshal(configKey, in)
	if err != nil {
		return result, err
	}

	for _, f := range b {
		if err := f(result); err != nil {
			return result, err
		}
	}

	return result, nil
}

// Optional is similar to run, but permits the configuration key to be missing.  In that event,
// no server is started and an information log message is output.
func Optional(configKey string, in ServerIn, b ...RunBuilder) (UnmarshalResult, error) {
	result, err := Run(configKey, in, b...)
	if _, ok := err.(ServerNotConfiguredError); ok {
		in.Logger.Log(
			level.Key(), level.InfoValue(),
			xlog.MessageKey(), "server not configured",
			ServerKey(), result.Name,
		)

		return result, nil
	}

	return result, err
}
