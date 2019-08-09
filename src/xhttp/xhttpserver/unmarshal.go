package xhttpserver

import (
	"config"
	"xlog"
	"xlog/xloghttp"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"go.uber.org/fx"
)

// ServerIn holds the set of dependencies required to create an HTTP server in the context
// of a uber/fx application.
//
// This struct is typically embedded in other fx.In structs so that Unmarshal can be invoked.
type ServerIn struct {
	fx.In

	Logger            log.Logger
	Unmarshaller      config.Unmarshaller
	Shutdowner        fx.Shutdowner
	Lifecycle         fx.Lifecycle
	ParameterBuilders xloghttp.ParameterBuilders `optional:"true"`
}

func unmarshal(configKey string, in ServerIn) (*mux.Router, error) {
	var o Options
	if err := config.UnmarshalRequired(in.Unmarshaller, configKey, &o); err != nil {
		return nil, err
	}

	if len(o.Name) == 0 {
		o.Name = configKey
	}

	var (
		serverLogger = NewServerLogger(o, in.Logger)
		serverChain  = NewServerChain(o, serverLogger, in.ParameterBuilders...)
		router       = mux.NewRouter()
		server       = New(o, serverLogger, serverChain.Then(router))
	)

	in.Lifecycle.Append(fx.Hook{
		OnStart: OnStart(serverLogger, server, func() { in.Shutdowner.Shutdown() }, o),
		OnStop:  OnStop(serverLogger, server),
	})

	return router, nil
}

func Required(configKey string) fx.Annotated {
	return fx.Annotated{
		Name: configKey,
		Target: func(in ServerIn) (*mux.Router, error) {
			return unmarshal(configKey, in)
		},
	}
}

func Optional(configKey string) fx.Annotated {
	return fx.Annotated{
		Name: configKey,
		Target: func(in ServerIn) (*mux.Router, error) {
			r, err := unmarshal(configKey, in)
			if _, ok := err.(config.MissingKeyError); ok {
				in.Logger.Log(
					level.Key(), level.InfoValue(),
					"configKey", configKey,
					xlog.MessageKey(), "server not configured",
				)

				return nil, nil
			}

			return r, err
		},
	}
}
