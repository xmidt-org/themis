package xmetrics

import (
	"github.com/go-kit/kit/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
)

func ProvideCounter(o prometheus.CounterOpts, labelNames ...string) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(r Registry) (metrics.Counter, error) {
			return r.NewCounter(o, labelNames)
		},
	}
}

func ProvideGauge(o prometheus.GaugeOpts, labelNames ...string) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(r Registry) (metrics.Gauge, error) {
			return r.NewGauge(o, labelNames)
		},
	}
}

func ProvideHistogram(o prometheus.HistogramOpts, labelNames ...string) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(r Registry) (metrics.Histogram, error) {
			return r.NewHistogram(o, labelNames)
		},
	}
}

func ProvideSummary(o prometheus.SummaryOpts, labelNames ...string) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(r Registry) (metrics.Histogram, error) {
			return r.NewSummary(o, labelNames)
		},
	}
}
