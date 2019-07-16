package xhealth

import (
	health "github.com/InVisionApp/go-health"
	"github.com/go-kit/kit/log"
	"github.com/spf13/viper"

	"go.uber.org/fx"
)

type HealthIn struct {
	fx.In

	Logger         log.Logger
	Viper          *viper.Viper
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
		if err := in.Viper.UnmarshalKey(configKey, &o); err != nil {
			return HealthOut{}, err
		}

		h, err := New(in.Logger, in.StatusListener, o)
		if err != nil {
			return HealthOut{}, err
		}

		in.Lifecycle.Append(fx.Hook{
			OnStart: OnStart(h),
			OnStop:  OnStop(h),
		})

		return HealthOut{
			Health:  h,
			Handler: NewHandler(h, o.Custom),
		}, nil
	}
}
