package xhttpserver

import (
	"strings"

	"github.com/xmidt-org/themis/config"
	"github.com/xmidt-org/themis/xlog/xloghttp"

	"github.com/go-kit/kit/log"
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

	Logger       log.Logger
	Unmarshaller config.Unmarshaller
	Shutdowner   fx.Shutdowner
	Lifecycle    fx.Lifecycle

	// ChainFactory is an optional component which is used to build an alice.Chain for each particular
	// server based on configuration.  Both this field and Chain may be used simultaneously.
	ChainFactory ChainFactory `optional:"true"`

	// ParameterBuiders is an optional component which is used to create contextual request loggers
	// for use by http.Handler code.
	ParameterBuilders xloghttp.ParameterBuilders `optional:"true"`
}

// Unmarshal unmarshals a server from the given configuration key and emits a *mux.Router.
//
// This function provides a default server name if none is supplied in the options.  This default name
// is the last dotted element of the configuration key.  This allows multiple servers in a map to naturally
// take their map keys as default names.  For example, unmarshalling a key of "servers.foo" would yield
// a default name of "foo".
func Unmarshal(configKey string, c ...alice.Constructor) func(in ServerIn) (*mux.Router, error) {
	return func(in ServerIn) (*mux.Router, error) {
		var o Options
		if err := in.Unmarshaller.UnmarshalKey(configKey, &o); err != nil {
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
			serverChain  = NewServerChain(o, serverLogger, in.ParameterBuilders...)
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
}
