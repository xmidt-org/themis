package xmetrics

import (
	"xhttp/xhttpserver"

	"github.com/go-kit/kit/log"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type InvokeIn struct {
	fx.In

	Logger     log.Logger
	Viper      *viper.Viper
	Shutdowner fx.Shutdowner
	Lifecycle  fx.Lifecycle

	Handler Handler
}

func InvokeServer(configKey string, path string) func(InvokeIn) error {
	return func(in InvokeIn) error {
		r, err := xhttpserver.Provide(configKey)(
			xhttpserver.ProvideIn{
				Logger:     in.Logger,
				Viper:      in.Viper,
				Shutdowner: in.Shutdowner,
				Lifecycle:  in.Lifecycle,
			},
		)

		if err != nil {
			return err
		}

		r.Handle(path, in.Handler)
		return nil
	}
}
