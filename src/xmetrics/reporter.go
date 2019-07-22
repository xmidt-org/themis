package xmetrics

import (
	"github.com/go-kit/kit/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

// Reporter encapsulates the notion of a metric
type Reporter interface {
	// Report uses the supplied labels/values to report a given value
	Report(*Labels, float64)
}

type ReporterFunc func(*Labels, float64)

func (rf ReporterFunc) Report(l *Labels, value float64) {
	rf(l, value)
}

type DiscardReporter struct{}

func (dr DiscardReporter) Report(*Labels, float64) {
}

// AdderReporter is a Reporter that expects positive values only
type AdderReporter Reporter

// SetterReporter is a Reporter that expects values to be the current value of the underlying metric
type SetterReporter Reporter

// ObserverReporter is a Reporter that expects values to be observations in a sequence
type ObserverReporter Reporter

func NewCounterReporter(c metrics.Counter) AdderReporter {
	return ReporterFunc(func(l *Labels, delta float64) {
		c.With(l.NamesAndValues()...).Add(delta)
	})
}

func NewCounterVecReporter(c *prometheus.CounterVec) AdderReporter {
	return ReporterFunc(func(l *Labels, delta float64) {
		c.WithLabelValues(l.Values()...).Add(delta)
	})
}

func NewGaugeAdderReporter(g metrics.Gauge) AdderReporter {
	return ReporterFunc(func(l *Labels, delta float64) {
		g.With(l.NamesAndValues()...).Add(delta)
	})
}

func NewGaugeVecAdderReporter(g *prometheus.GaugeVec) AdderReporter {
	return ReporterFunc(func(l *Labels, delta float64) {
		g.WithLabelValues(l.Values()...).Add(delta)
	})
}

func NewGaugeSetterReporter(g metrics.Gauge) SetterReporter {
	return ReporterFunc(func(l *Labels, delta float64) {
		g.With(l.NamesAndValues()...).Add(delta)
	})
}

func NewGaugeVecSetterReporter(g *prometheus.GaugeVec) SetterReporter {
	return ReporterFunc(func(l *Labels, delta float64) {
		g.WithLabelValues(l.Values()...).Add(delta)
	})
}

func NewHistogramReporter(h metrics.Histogram) ObserverReporter {
	return ReporterFunc(func(l *Labels, value float64) {
		h.With(l.NamesAndValues()...).Observe(value)
	})
}

func NewObserverVecReporter(o prometheus.ObserverVec) ObserverReporter {
	return ReporterFunc(func(l *Labels, value float64) {
		o.WithLabelValues(l.Values()...).Observe(value)
	})
}
