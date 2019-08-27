package xmetrics

import (
	"github.com/xmidt-org/themis/config"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
)

type MetricsIn struct {
	fx.In

	Unmarshaller config.Unmarshaller
}

type MetricsOut struct {
	fx.Out

	Registerer prometheus.Registerer
	Gatherer   prometheus.Gatherer
	Factory    Factory
	Registry   Registry
}

// Unmarshal produces an uber/fx provider that bootstraps a prometheus-based metrics environment.
// No HTTP initialization is done by this package.  To obtain a prometheus handler, use xmetricshttp.Unmarshal,
// which invokes this method in addition to the HTTP initialization.
func Unmarshal(configKey string) func(MetricsIn) (MetricsOut, error) {
	return func(in MetricsIn) (MetricsOut, error) {
		var o Options
		if err := in.Unmarshaller.UnmarshalKey(configKey, &o); err != nil {
			return MetricsOut{}, err
		}

		registry, err := New(o)
		if err != nil {
			return MetricsOut{}, err
		}

		return MetricsOut{
			Registerer: registry,
			Gatherer:   registry,
			Factory:    registry,
			Registry:   registry,
		}, nil
	}
}
