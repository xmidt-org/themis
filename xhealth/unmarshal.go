package xhealth

import (
	"github.com/xmidt-org/themis/config"

	health "github.com/InVisionApp/go-health"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// HealthIn defines the set of dependencies for instantiating an InVision health service
// and binding that service to the application lifecycle.
type HealthIn struct {
	fx.In

	// Logger is the required go-kit logger that will receive health logging output
	logger *zap.Logger

	// Unmarshaller is the required configuration unmarshaller strategy
	Unmarshaller config.Unmarshaller

	// Lifecycle is used to bind the health service the the uber/fx App lifecycle
	Lifecycle fx.Lifecycle

	// StatusListener is the optional listener for health status changes
	StatusListener health.IStatusListener `optional:"true"`

	// Config is an optional check.  If both this field and Configs are set, both fields
	// are added.
	Config *health.Config `optional:"true"`

	// Configs is an optional slice of checks.  If both this field and Config are set, both
	// fields are added.
	Configs []*health.Config `optional:"true"`
}

// HealthOut defines the components emitted by this package
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

		if in.Config != nil {
			if err := h.AddCheck(in.Config); err != nil {
				return HealthOut{}, err
			}
		}

		if len(in.Configs) > 0 {
			if err := h.AddChecks(in.Configs); err != nil {
				return HealthOut{}, err
			}
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
