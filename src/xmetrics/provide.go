package xmetrics

import (
	"github.com/go-kit/kit/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
)

// ProvideCounter emits an uber/fx component of type metrics.Counter using the unqualified name specified
// in the options struct.
func ProvideCounter(o prometheus.CounterOpts, labelNames ...string) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(r Registry) (metrics.Counter, error) {
			return r.NewCounter(o, labelNames)
		},
	}
}

// ProvideGauge emits an uber/fx component of type metrics.Gauge using the unqualified name specified
// in the options struct.
func ProvideGauge(o prometheus.GaugeOpts, labelNames ...string) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(r Registry) (metrics.Gauge, error) {
			return r.NewGauge(o, labelNames)
		},
	}
}

// ProvideHistogram emits an uber/fx component of type metrics.Histogram using the unqualified name specified
// in the options struct.
func ProvideHistogram(o prometheus.HistogramOpts, labelNames ...string) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(r Registry) (metrics.Histogram, error) {
			return r.NewHistogram(o, labelNames)
		},
	}
}

// ProvideHistogram emits an uber/fx component of type metrics.Histogram using the unqualified name specified
// in the options struct.  Note that go-kit does not have a separate summary metric type.
func ProvideSummary(o prometheus.SummaryOpts, labelNames ...string) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(r Registry) (metrics.Histogram, error) {
			return r.NewSummary(o, labelNames)
		},
	}
}
