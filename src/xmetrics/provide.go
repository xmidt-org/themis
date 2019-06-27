package xmetrics

import (
	"github.com/go-kit/kit/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type ProvideOut struct {
	fx.Out

	Registerer prometheus.Registerer
	Gatherer   prometheus.Gatherer
	Provider   Provider
}

func Provide(configKey string) func(*viper.Viper) (ProvideOut, error) {
	return func(v *viper.Viper) (ProvideOut, error) {
		var o Options
		if err := v.UnmarshalKey(configKey, &o); err != nil {
			return ProvideOut{}, err
		}

		p, err := New(o)
		if err != nil {
			return ProvideOut{}, err
		}

		return ProvideOut{
			Registerer: p,
			Gatherer:   p,
			Provider:   p,
		}, nil
	}
}

func ProvideHandler(o promhttp.HandlerOpts) func(prometheus.Gatherer) Handler {
	return func(g prometheus.Gatherer) Handler {
		return NewHandler(g, o)
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
