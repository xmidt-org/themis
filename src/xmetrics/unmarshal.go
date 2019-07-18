package xmetrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type MetricsIn struct {
	fx.In

	Viper *viper.Viper
}

type MetricsOut struct {
	fx.Out

	Registerer prometheus.Registerer
	Gatherer   prometheus.Gatherer
	Registry   Registry
}

// Unmarshal produces an uber/fx provider that bootstraps a prometheus-based metrics environment.
// No HTTP initialization is done by this package.  To obtain a prometheus handler, use xmetricshttp.Unmarshal,
// which invokes this method in addition to the HTTP initialization.
func Unmarshal(configKey string) func(MetricsIn) (MetricsOut, error) {
	return func(in MetricsIn) (MetricsOut, error) {
		var o Options
		if err := in.Viper.UnmarshalKey(configKey, &o); err != nil {
			return MetricsOut{}, err
		}

		registry, err := New(o)
		if err != nil {
			return MetricsOut{}, err
		}

		return MetricsOut{
			Registerer: registry,
			Gatherer:   registry,
			Registry:   registry,
		}, nil
	}
}
