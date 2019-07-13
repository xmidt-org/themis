package xhttpserver

import (
	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

// ServerIn holds the set of dependencies required to create an HTTP server in the context
// of a uber/fx application.
//
// This struct is typically embedded in other fx.In structs so that Unmarshal can be invoked.
type ServerIn struct {
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
// This function is useful for writing server invocation code for other packages, typically the main package.
func Unmarshal(configKey string, in ServerIn) (*mux.Router, log.Logger, error) {
	var o Options
	if err := in.Viper.UnmarshalKey(configKey, &o); err != nil {
		return nil, nil, err
	}

	if len(o.Name) == 0 {
		o.Name = configKey
	}

	router := mux.NewRouter()
	server, logger, err := New(in.Logger, router, o)
	if err != nil {
		return nil, nil, err
	}

	in.Lifecycle.Append(fx.Hook{
		OnStart: OnStart(logger, server, func() { in.Shutdowner.Shutdown() }, o),
		OnStop:  OnStop(logger, server),
	})

	return router, logger, nil
}
