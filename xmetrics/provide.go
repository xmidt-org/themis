// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
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
		Target: func(f Factory) (metrics.Counter, error) {
			return f.NewCounter(o, labelNames)
		},
	}
}

// ProvideCounterVec emits an uber/fx component of the *prometheus.CounterVec using the unqualified name
// specified in the options struct.  Use this provider when lower-level access to prometheus features is needed.
func ProvideCounterVec(o prometheus.CounterOpts, labelNames ...string) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(f Factory) (*prometheus.CounterVec, error) {
			return f.NewCounterVec(o, labelNames)
		},
	}
}

// ProvideGauge emits an uber/fx component of type metrics.Gauge using the unqualified name specified
// in the options struct.
func ProvideGauge(o prometheus.GaugeOpts, labelNames ...string) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(f Factory) (metrics.Gauge, error) {
			return f.NewGauge(o, labelNames)
		},
	}
}

// ProvideGaugeVec emits an uber/fx component of the *prometheus.GaugeVec using the unqualified name
// specified in the options struct.  Use this provider when lower-level access to prometheus features is needed.
func ProvideGaugeVec(o prometheus.GaugeOpts, labelNames ...string) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(f Factory) (*prometheus.GaugeVec, error) {
			return f.NewGaugeVec(o, labelNames)
		},
	}
}

// ProvideHistogram emits an uber/fx component of type metrics.Histogram using the unqualified name specified
// in the options struct.
func ProvideHistogram(o prometheus.HistogramOpts, labelNames ...string) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(f Factory) (metrics.Histogram, error) {
			return f.NewHistogram(o, labelNames)
		},
	}
}

// ProvideHistogramVec emits an uber/fx component of the *prometheus.HistogramVec using the unqualified name
// specified in the options struct.  Use this provider when lower-level access to prometheus features is needed.
func ProvideHistogramVec(o prometheus.HistogramOpts, labelNames ...string) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(f Factory) (*prometheus.HistogramVec, error) {
			return f.NewHistogramVec(o, labelNames)
		},
	}
}

// ProvideHistogram emits an uber/fx component of type metrics.Histogram using the unqualified name specified
// in the options struct.  Note that go-kit does not have a separate summary metric type.
func ProvideSummary(o prometheus.SummaryOpts, labelNames ...string) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(f Factory) (metrics.Histogram, error) {
			return f.NewSummary(o, labelNames)
		},
	}
}

// ProvideSummaryVec emits an uber/fx component of the *prometheus.SummaryVec using the unqualified name
// specified in the options struct.  Use this provider when lower-level access to prometheus features is needed.
func ProvideSummaryVec(o prometheus.SummaryOpts, labelNames ...string) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(f Factory) (*prometheus.SummaryVec, error) {
			return f.NewSummaryVec(o, labelNames)
		},
	}
}
