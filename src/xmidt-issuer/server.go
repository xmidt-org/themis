package main

import (
	"net/http"
	"xhttp/xhttpserver"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type MainIn struct {
	fx.In

	Logger     log.Logger
	Viper      *viper.Viper
	Shutdowner fx.Shutdowner
	Lifecycle  fx.Lifecycle
}

func ProvideMain(serverKey string) func(MainIn) (*mux.Router, error) {
	return func(in MainIn) (*mux.Router, error) {
		var o xhttpserver.Options
		if err := in.Viper.UnmarshalKey(serverKey, &o); err != nil {
			return nil, err
		}

		router := mux.NewRouter()
		server, logger, err := xhttpserver.New(in.Logger, router, o)
		if err != nil {
			return nil, err
		}

		in.Lifecycle.Append(fx.Hook{
			OnStart: xhttpserver.OnStart(logger, server, func() { in.Shutdowner.Shutdown() }, o),
			OnStop:  xhttpserver.OnStop(logger, server),
		})

		return router, nil
	}
}

type RouteIn struct {
	fx.In

	Router       *mux.Router
	KeyHandler   http.Handler `name:"keyHandler"`
	IssueHandler http.Handler `name:"issueHandler"`
}

func DefineMainRoutes(in RouteIn) {
	in.Router.Handle("/issue", in.IssueHandler).Methods("GET")
	in.Router.Handle("/keys/{kid}", in.KeyHandler).Methods("GET")
}
