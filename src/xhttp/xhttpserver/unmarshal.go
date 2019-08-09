package xhttpserver

import (
	"config"
	"strings"
	"xlog"
	"xlog/xloghttp"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"go.uber.org/fx"
)

// ChainFactory is a creation strategy for server-specific alice.Chains that will decorate the
// server handler.  Chains created by this factory will be appended to the core chain created
// by NewServerChain.
//
// This interface is useful when particular servers need custom chains based on configuration.
// The most common example of this is metrics, as server metrics might need the name of the
// server as a label.
type ChainFactory interface {
	New(Options) (alice.Chain, error)
}

type ChainFactoryFunc func(Options) (alice.Chain, error)

func (cff ChainFactoryFunc) New(o Options) (alice.Chain, error) {
	return cff(o)
}

// ServerIn holds the set of dependencies required to create an HTTP server in the context
// of a uber/fx application.
type ServerIn struct {
	fx.In

	Logger            log.Logger
	Unmarshaller      config.Unmarshaller
	Shutdowner        fx.Shutdowner
	Lifecycle         fx.Lifecycle
	Chain             alice.Chain                `optional:"true"`
	ChainFactory      ChainFactory               `optional:"true"`
	ParameterBuilders xloghttp.ParameterBuilders `optional:"true"`
}

// unmarshal implements the common, internal logic to provide a server
func unmarshal(configKey string, in ServerIn, c ...alice.Constructor) (*mux.Router, error) {
	var o Options
	if err := config.UnmarshalRequired(in.Unmarshaller, configKey, &o); err != nil {
		return nil, err
	}

	if len(o.Name) == 0 {
		if pos := strings.LastIndexByte(configKey, '.'); pos >= 0 {
			o.Name = configKey[pos+1:]
		} else {
			o.Name = configKey
		}
	}

	var (
		serverLogger = NewServerLogger(o, in.Logger)
		serverChain  = NewServerChain(o, serverLogger, in.ParameterBuilders...).Extend(in.Chain)
	)

	if in.ChainFactory != nil {
		more, err := in.ChainFactory.New(o)
		if err != nil {
			return nil, err
		}

		serverChain = serverChain.Extend(more)
	}

	var (
		router = mux.NewRouter()
		server = New(
			o,
			serverLogger,
			serverChain.Append(c...).Then(router),
		)
	)

	in.Lifecycle.Append(fx.Hook{
		OnStart: OnStart(o, server, serverLogger, func() { in.Shutdowner.Shutdown() }),
		OnStop:  OnStop(server, serverLogger),
	})

	return router, nil
}

// Required unmarshals a server from the given configuration key and emits a *mux.Router.
// This provider raises an error if the configuration key does not exist.
//
// This function provides a default server name if none is supplied in the options.  This default name
// is the last dotted element of the configuration key.  This allows multiple servers in a map to naturally
// take their map keys as default names.  For example, unmarshalling a key of "servers.foo" would yield
// a default name of "foo".
func Required(configKey string, c ...alice.Constructor) func(in ServerIn) (*mux.Router, error) {
	return func(in ServerIn) (*mux.Router, error) {
		return unmarshal(configKey, in, c...)
	}
}

// Optional unmarshals a server from the given configuration key, returning a nil *mux.Router if
// no such configuration key is found.  In all other ways, this function is the same as Required.
func Optional(configKey string, c ...alice.Constructor) func(in ServerIn) (*mux.Router, error) {
	return func(in ServerIn) (*mux.Router, error) {
		r, err := unmarshal(configKey, in, c...)
		if _, ok := err.(config.MissingKeyError); ok {
			in.Logger.Log(
				level.Key(), level.InfoValue(),
				"configKey", configKey,
				xlog.MessageKey(), "server not configured",
			)

			return nil, nil
		}

		return r, err
	}
}
