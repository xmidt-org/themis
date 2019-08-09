package xhealth

import (
	"config"

	health "github.com/InVisionApp/go-health"
	"github.com/go-kit/kit/log"

	"go.uber.org/fx"
)

type HealthIn struct {
	fx.In

	Logger         log.Logger
	Unmarshaller   config.Unmarshaller
	Lifecycle      fx.Lifecycle
	StatusListener health.IStatusListener `optional:"true"`
}

type HealthOut struct {
	fx.Out

	Health  health.IHealth
	Handler Handler
}

// Unmarshal returns an uber/fx provider that reads configuration from a Viper
// instance and initializes the health infrastructure.
func Unmarshal(configKey string) func(HealthIn) (HealthOut, error) {
	return func(in HealthIn) (HealthOut, error) {
		var o Options
		if err := in.Unmarshaller.UnmarshalKey(configKey, &o); err != nil {
			return HealthOut{}, err
		}

		h, err := New(o, in.Logger, in.StatusListener)
		if err != nil {
			return HealthOut{}, err
		}

		in.Lifecycle.Append(fx.Hook{
			OnStart: OnStart(in.Logger, h),
			OnStop:  OnStop(in.Logger, h),
		})

		return HealthOut{
			Health:  h,
			Handler: NewHandler(h, o.Custom),
		}, nil
	}
}
