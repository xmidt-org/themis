package xmetrics

import (
	"github.com/go-kit/kit/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

func Provide(configKey string) func(*viper.Viper) (prometheus.Registerer, prometheus.Gatherer, Provider, error) {
	return func(v *viper.Viper) (prometheus.Registerer, prometheus.Gatherer, Provider, error) {
		var o Options
		if err := v.UnmarshalKey(configKey, &o); err != nil {
			return nil, nil, nil, err
		}

		p, err := New(o)
		if err != nil {
			return nil, nil, nil, err
		}

		return p, p, p, nil
	}
}

func ProvideCounter(o prometheus.CounterOpts, labelNames ...string) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(p Provider) (metrics.Counter, error) {
			return p.NewCounter(o, labelNames)
		},
	}
}

func ProvideGauge(o prometheus.GaugeOpts, labelNames ...string) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(p Provider) (metrics.Gauge, error) {
			return p.NewGauge(o, labelNames)
		},
	}
}

func ProvideHistogram(o prometheus.HistogramOpts, labelNames ...string) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(p Provider) (metrics.Histogram, error) {
			return p.NewHistogram(o, labelNames)
		},
	}
}

func ProvideSummary(o prometheus.SummaryOpts, labelNames ...string) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(p Provider) (metrics.Histogram, error) {
			return p.NewSummary(o, labelNames)
		},
	}
}
