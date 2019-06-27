package xhttpserver

import (
	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

// ProvideIn holds the set of dependencies required to create an HTTP server in the context
// of a uber/fx application.
type ProvideIn struct {
	fx.In

	Logger     log.Logger
	Viper      *viper.Viper
	Shutdowner fx.Shutdowner
	Lifecycle  fx.Lifecycle
}

// Provide returns a closure that unmarshals server options, creates the server, binds to the fx App lifecycle,
// and then returns a *mux.Router than can be used to configure the handler routes for the server.  This one-stop
// shopping function serves most use cases.
func Provide(configKey string) func(ProvideIn) (*mux.Router, error) {
	return func(in ProvideIn) (*mux.Router, error) {
		var o Options
		if err := in.Viper.UnmarshalKey(configKey, &o); err != nil {
			return nil, err
		}

		router := mux.NewRouter()
		server, logger, err := New(in.Logger, router, o)
		if err != nil {
			return nil, err
		}

		in.Lifecycle.Append(fx.Hook{
			OnStart: OnStart(logger, server, func() { in.Shutdowner.Shutdown() }, o),
			OnStop:  OnStop(logger, server),
		})

		return router, nil
	}
}
