package xhttpserver

import (
	"fmt"

	"github.com/xmidt-org/themis/config"
	"github.com/xmidt-org/themis/xlog/xloghttp"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"go.uber.org/fx"
)

// ServerNotConfiguredError is returned when a required server has no configuration key
type ServerNotConfiguredError struct {
	Key string
}

func (e ServerNotConfiguredError) Error() string {
	return fmt.Sprintf("No server with key %s is configured.", e.Key)
}

// ChainFactory is a creation strategy for server-specific alice.Chains that will decorate the
// server handler.  Chains created by this factory will be appended to the core chain created
// by NewServerChain.
//
// This interface is useful when particular servers need custom chains based on configuration.
// The most common example of this is metrics, as server metrics might need the name of the
// server as a label.
type ChainFactory interface {
	New(string, Options) (alice.Chain, error)
}

type ChainFactoryFunc func(string, Options) (alice.Chain, error)

func (cff ChainFactoryFunc) New(n string, o Options) (alice.Chain, error) {
	return cff(n, o)
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

// Unmarshal describes how to unmarshal an HTTP server.  This type contains all the non-component information
// related to server instantiation.
type Unmarshal struct {
	// Key is the viper configuration key containing the server Options
	Key string

	// Name is the string that identifies this server from others within the same application.  If unset,
	// the Key is used.
	Name string

	// Optional indicates whether the configuration is required.  If this field is false (the default),
	// and there is no such configuration Key, an error is returned.
	Optional bool

	// Chain is an optional set of constructors that will decorate the *mux.Router.  This field is useful for static
	// decorators, such as inserting known headers into every response.
	//
	// This chain cannot depend on components.  In order to leverage dependency injection, create a ChainFactory instead.
	Chain alice.Chain
}

func (u Unmarshal) name() string {
	if len(u.Name) > 0 {
		return u.Name
	}

	return u.Key
}

// Provide unmarshals a server using the Key field and creates a *mux.Router which is the root handler for
// that server's requests.  This *mux.Router will be decorated with the constructors from NewServerChain as well
// as any ChainFactory's constructors.
func (u Unmarshal) Provide(in ServerIn) (*mux.Router, error) {
	if !in.Unmarshaller.IsSet(u.Key) {
		if !u.Optional {
			return nil, ServerNotConfiguredError{Key: u.Key}
		}

		return nil, nil
	}

	var o Options
	if err := in.Unmarshaller.UnmarshalKey(u.Key, &o); err != nil {
		return nil, err
	}

	var (
		serverName   = u.name()
		serverLogger = log.With(in.Logger, ServerKey(), serverName)
		serverChain  = NewServerChain(o, serverLogger, in.ParameterBuilders...)
	)

	if in.ChainFactory != nil {
		more, err := in.ChainFactory.New(serverName, o)
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
			serverChain.Extend(u.Chain).Then(router),
		)
	)

	in.Lifecycle.Append(fx.Hook{
		OnStart: OnStart(o, server, serverLogger, func() { in.Shutdowner.Shutdown() }),
		OnStop:  OnStop(server, serverLogger),
	})

	return router, nil
}

// Annotated is like Unmarshal, save that it emits a named *mux.Router.  This method is appropriate
// for applications with multiple servers.  The name of the returned *mux.Router is either the Name field (if set)
// or the Key field (if Name is empty).
func (u Unmarshal) Annotated() fx.Annotated {
	return fx.Annotated{
		Name:   u.name(),
		Target: u.Provide,
	}
}
