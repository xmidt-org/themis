package xhttpserver

import (
	"fmt"
	"strings"
	"xconfig"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type ServerNotConfiguredError struct {
	ConfigKey string
}

func (snce ServerNotConfiguredError) Error() string {
	return fmt.Sprintf("No server configured with key %s", snce.ConfigKey)
}

// ServerIn holds the set of dependencies required to create an HTTP server in the context
// of a uber/fx application.
//
// This struct is typically embedded in other fx.In structs so that Unmarshal can be invoked.
type ServerIn struct {
	fx.In

	Logger       log.Logger
	Viper        *viper.Viper
	Unmarshaller xconfig.KeyUnmarshaller
	Shutdowner   fx.Shutdowner
	Lifecycle    fx.Lifecycle
}

// UnmarshalResult is the result of unmarshalling a server and binding it to the container lifecycle
type UnmarshalResult struct {
	// Name is the label applied to this server in logging.  It will either be set via configuration
	// or default to the configuration key.
	Name string

	// Logger is the go-kit logger enriched with server information, such as the bind address
	Logger log.Logger

	// Router is the gorilla/mux router used as the handler for this server, which can be used
	// to build handler routes.
	Router *mux.Router
}

// Unmarshal reads an Options struct at the given viper key, creates an HTTP server instance,
// binds it to the fx.App lifecycle, and returns a gorilla/mux router that can be used to
// define handler routes for the server.
//
// This function is useful for writing server invocation code for other packages, typically the main package.
// It is not intended for direct use as an uber/fx provider.
//
// Even when returning an error, this function always returns an UnmarshalResult with at least the server name
// set to something that can be output for information and debugging.
func Unmarshal(configKey string, in ServerIn) (UnmarshalResult, error) {
	if !in.Viper.IsSet(configKey) {
		return UnmarshalResult{Name: configKey}, ServerNotConfiguredError{ConfigKey: configKey}
	}

	var o Options
	if err := in.Unmarshaller.UnmarshalKey(configKey, &o); err != nil {
		return UnmarshalResult{Name: configKey}, err
	}

	if len(o.Name) == 0 {
		defaultName := configKey
		if i := strings.LastIndexByte(defaultName, '.'); i >= 0 {
			defaultName = configKey[i+1:]
		}

		o.Name = defaultName
	}

	router := mux.NewRouter()
	server, logger := New(in.Logger, router, o)

	in.Lifecycle.Append(fx.Hook{
		OnStart: OnStart(logger, server, func() { in.Shutdowner.Shutdown() }, o),
		OnStop:  OnStop(logger, server),
	})

	return UnmarshalResult{
		Name:   o.Name,
		Logger: logger,
		Router: router,
	}, nil
}
