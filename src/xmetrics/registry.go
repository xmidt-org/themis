package xmetrics

import (
	"github.com/go-kit/kit/metrics"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
)

// Adder is a strategy for adding a value to a metric.  This interface is implemented
// by a variety of metrics infrastructures.
type Adder interface {
	Add(float64)
}

// Setter is a strategy for setting a value on a metric.  This interface is implemented
// by a variety of metrics infrastructures.
type Setter interface {
	Set(float64)
}

// Observer is a strategy for observing a value on a metric.  This interface is implemented
// by a variety of metrics infrastructures.
type Observer interface {
	Observe(float64)
}

type Options struct {
	DefaultNamespace        string
	DefaultSubsystem        string
	Pedantic                bool
	DisableGoCollector      bool
	DisableProcessCollector bool
}

type Registry interface {
	prometheus.Registerer
	prometheus.Gatherer

	NewCounter(prometheus.CounterOpts, []string) (metrics.Counter, error)
	NewGauge(prometheus.GaugeOpts, []string) (metrics.Gauge, error)
	NewHistogram(prometheus.HistogramOpts, []string) (metrics.Histogram, error)
	NewSummary(prometheus.SummaryOpts, []string) (metrics.Histogram, error)
}

type registry struct {
	prometheus.Registerer
	prometheus.Gatherer

	defaultNamespace string
	defaultSubsystem string
}

func (r *registry) namespace(v string) string {
	if len(v) > 0 {
		return v
	}

	return r.defaultNamespace
}

func (r *registry) subsystem(v string) string {
	if len(v) > 0 {
		return v
	}

	return r.defaultSubsystem
}

func (r *registry) NewCounter(o prometheus.CounterOpts, labelNames []string) (metrics.Counter, error) {
	o.Namespace = r.namespace(o.Namespace)
	o.Subsystem = r.subsystem(o.Subsystem)

	cv := prometheus.NewCounterVec(o, labelNames)
	if err := r.Register(cv); err != nil {
		return nil, err
	}

	return kitprometheus.NewCounter(cv), nil
}

func (r *registry) NewGauge(o prometheus.GaugeOpts, labelNames []string) (metrics.Gauge, error) {
	o.Namespace = r.namespace(o.Namespace)
	o.Subsystem = r.subsystem(o.Subsystem)

	cv := prometheus.NewGaugeVec(o, labelNames)
	if err := r.Register(cv); err != nil {
		return nil, err
	}

	return kitprometheus.NewGauge(cv), nil
}

func (r *registry) NewHistogram(o prometheus.HistogramOpts, labelNames []string) (metrics.Histogram, error) {
	o.Namespace = r.namespace(o.Namespace)
	o.Subsystem = r.subsystem(o.Subsystem)

	cv := prometheus.NewHistogramVec(o, labelNames)
	if err := r.Register(cv); err != nil {
		return nil, err
	}

	return kitprometheus.NewHistogram(cv), nil
}

func (r *registry) NewSummary(o prometheus.SummaryOpts, labelNames []string) (metrics.Histogram, error) {
	o.Namespace = r.namespace(o.Namespace)
	o.Subsystem = r.subsystem(o.Subsystem)

	cv := prometheus.NewSummaryVec(o, labelNames)
	if err := r.Register(cv); err != nil {
		return nil, err
	}

	return kitprometheus.NewSummary(cv), nil
}

func New(o Options) (Registry, error) {
	var pr *prometheus.Registry
	if o.Pedantic {
		pr = prometheus.NewRegistry()
	} else {
		pr = prometheus.NewPedanticRegistry()
	}

	if !o.DisableGoCollector {
		if err := pr.Register(prometheus.NewGoCollector()); err != nil {
			return nil, err
		}
	}

	if !o.DisableProcessCollector {
		var pco prometheus.ProcessCollectorOpts
		if err := pr.Register(prometheus.NewProcessCollector(pco)); err != nil {
			return nil, err
		}
	}

	return &registry{
		Registerer:       pr,
		Gatherer:         pr,
		defaultNamespace: o.DefaultNamespace,
		defaultSubsystem: o.DefaultSubsystem,
	}, nil
}
