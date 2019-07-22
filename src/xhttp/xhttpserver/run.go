package xhttpserver

import (
	"net/http/pprof"
	"xlog"

	"github.com/go-kit/kit/log/level"
)

// RunBuilder is a strategy for building up routes and other infrastructure related to an http.Server
type RunBuilder func(UnmarshalResult) error

// AddPprofRoutes is a RunBuilder that adds the same routes as the net/http/pprof package.  This builder
// can be used with Invoke, for example.
func AddPprofRoutes(ur UnmarshalResult) error {
	ur.Router.HandleFunc("/debug/pprof/", pprof.Index)
	ur.Router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	ur.Router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	ur.Router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	ur.Router.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return nil
}

// Run handles building an http.Server, using Unmarshal to load the server from configuration and
// then applying zero or more builders.  This function is not intended to be used directly as an fx.Invoke option.
// Rather, an invoke function can use Run for its heavylifting together with additional dependencies to be used
// in the builders.
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

// Invoke can be used directly as an fx.Invoke option.  This function is useful when no other
// dependencies are required for building the server.
func Invoke(configKey string, b ...RunBuilder) func(ServerIn) error {
	return func(in ServerIn) error {
		_, err := Run(configKey, in, b...)
		return err
	}
}

// InvokeOptional is like Invoke, but behaves like Optional when the server configuration is missing
func InvokeOptional(configKey string, b ...RunBuilder) func(ServerIn) error {
	return func(in ServerIn) error {
		_, err := Optional(configKey, in, b...)
		return err
	}
}
