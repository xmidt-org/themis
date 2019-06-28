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

// Unmarshal reads an Options struct at the given viper key, creates an HTTP server instance,
// binds it to the fx.App lifecycle, and returns a gorilla/mux router that can be used to
// define handler routes for the server.
//
// This function is useful when server creation needs to be embedded in other packages.
func Unmarshal(configKey string, in ProvideIn) (*mux.Router, error) {
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

// Provide returns a closure that invokes Unmarshal as part of an fx.Provide option.  This function
// is useful when directly providing a server to an fx.App.
func Provide(configKey string) func(ProvideIn) (*mux.Router, error) {
	return func(in ProvideIn) (*mux.Router, error) {
		return Unmarshal(configKey, in)
	}
}
