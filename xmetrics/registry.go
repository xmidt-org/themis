package xmetrics

import (
	"github.com/go-kit/kit/metrics"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
)

// Options defines the configuration options for bootstrapping a prometheus-based metrics environment
// within an uber/fx App backed by Viper configuration.
type Options struct {
	// DefaultNamespace is the prometheus namespace to apply when a metric has no namespace
	DefaultNamespace string

	// DefaultSubsystem is the prometheus subsystem to apply when a metric has no subsystem
	DefaultSubsystem string

	// Pedantic controls whether a pedantic Registerer is used as the prometheus backend.
	// See https://godoc.org/github.com/prometheus/client_golang/prometheus#NewPedanticRegistry
	Pedantic bool

	// DisableGoCollector controls whether the go collector is registered on startup.
	// By default, the go collector is registered.
	//
	// See https://godoc.org/github.com/prometheus/client_golang/prometheus#NewGoCollector
	DisableGoCollector bool

	// DisableProcessCollector controls whether the process collector is registered on startup.
	// By default, this collector is registered.
	//
	// See https://godoc.org/github.com/prometheus/client_golang/prometheus#NewProcessCollector
	DisableProcessCollector bool

	// ConstLabels is an optional map of constant labels and values that are applied to all
	// registered metrics.  Useful for defining application-wide metrics, usually to distinguish
	// running instances in a cluster.
	ConstLabels map[string]string
}

// Factory is a creational strategy go-kit and prometheus metrics
type Factory interface {
	// NewCounter constructs a go-kit Counter backed by a prometheus counter.  The wrapped counter
	// will be registered if this implementation is also a Registerer (the default).
	NewCounter(prometheus.CounterOpts, []string) (metrics.Counter, error)

	// NewCounterVec constructs a prometheus counter.  This counter will be registered if this implementation
	// is also a Registerer (the default).
	//
	// Use this method when lower level access to prometheus features are required, such as currying.
	NewCounterVec(prometheus.CounterOpts, []string) (*prometheus.CounterVec, error)

	// NewGauge constructs a go-kit Gauge backed by a prometheus gauge.  The wrapped gauge
	// will be registered if this implementation is also a Registerer (the default).
	NewGauge(prometheus.GaugeOpts, []string) (metrics.Gauge, error)

	// NewGaugeVec constructs a prometheus gauge.  This gauge will be registered if this implementation
	// is also a Registerer (the default).
	//
	// Use this method when lower level access to prometheus features are required, such as currying.
	NewGaugeVec(prometheus.GaugeOpts, []string) (*prometheus.GaugeVec, error)

	// NewHistogram constructs a go-kit Histogram backed by a prometheus histogram.  The wrapped histogram
	// will be registered if this implementation is also a Registerer (the default).
	NewHistogram(prometheus.HistogramOpts, []string) (metrics.Histogram, error)

	// NewHistogramVec constructs a prometheus histogram.  This histogram will be registered if this implementation
	// is also a Registerer (the default).
	//
	// Use this method when lower level access to prometheus features are required, such as currying.
	NewHistogramVec(prometheus.HistogramOpts, []string) (*prometheus.HistogramVec, error)

	// NewSummary constructs a go-kit Histogram backed by a prometheus summary.  The wrapped summary
	// will be registered if this implementation is also a Registerer (the default).
	//
	// Go-kit does not have a separate histogram vs summary interface.  Thus, client code wishing to stay
	// abstracted from prometheus needs to use go-kit's Histogram interface.
	NewSummary(prometheus.SummaryOpts, []string) (metrics.Histogram, error)

	// NewSummaryVec constructs a prometheus summary.  This summary will be registered if this implementation
	// is also a Registerer (the default).
	//
	// Use this method when lower level access to prometheus features are required, such as currying.
	NewSummaryVec(prometheus.SummaryOpts, []string) (*prometheus.SummaryVec, error)
}

// Registry is the central interface of this package.  It implements the appropriate prometheus interfaces
// and supplies factory methods that return go-kit metrics types.
type Registry interface {
	prometheus.Registerer
	prometheus.Gatherer
	Factory
}

type registry struct {
	prometheus.Registerer
	prometheus.Gatherer

	defaultNamespace string
	defaultSubsystem string
	constLabels      map[string]string
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

func (r *registry) mergeConstLabels(original prometheus.Labels) prometheus.Labels {
	copy := make(prometheus.Labels, len(original)+len(r.constLabels))
	for k, v := range original {
		copy[k] = v
	}

	for k, v := range r.constLabels {
		copy[k] = v
	}

	return copy
}

func (r *registry) NewCounter(o prometheus.CounterOpts, labelNames []string) (metrics.Counter, error) {
	cv, err := r.NewCounterVec(o, labelNames)
	if err != nil {
		return nil, err
	}

	return kitprometheus.NewCounter(cv), nil
}

func (r *registry) NewCounterVec(o prometheus.CounterOpts, labelNames []string) (*prometheus.CounterVec, error) {
	o.Namespace = r.namespace(o.Namespace)
	o.Subsystem = r.subsystem(o.Subsystem)
	o.ConstLabels = r.mergeConstLabels(o.ConstLabels)

	cv := prometheus.NewCounterVec(o, labelNames)
	if err := r.Register(cv); err != nil {
		return nil, err
	}

	return cv, nil
}

func (r *registry) NewGauge(o prometheus.GaugeOpts, labelNames []string) (metrics.Gauge, error) {
	gv, err := r.NewGaugeVec(o, labelNames)
	if err != nil {
		return nil, err
	}

	return kitprometheus.NewGauge(gv), nil
}

func (r *registry) NewGaugeVec(o prometheus.GaugeOpts, labelNames []string) (*prometheus.GaugeVec, error) {
	o.Namespace = r.namespace(o.Namespace)
	o.Subsystem = r.subsystem(o.Subsystem)
	o.ConstLabels = r.mergeConstLabels(o.ConstLabels)

	gv := prometheus.NewGaugeVec(o, labelNames)
	if err := r.Register(gv); err != nil {
		return nil, err
	}

	return gv, nil
}

func (r *registry) NewHistogram(o prometheus.HistogramOpts, labelNames []string) (metrics.Histogram, error) {
	hv, err := r.NewHistogramVec(o, labelNames)
	if err != nil {
		return nil, err
	}

	return kitprometheus.NewHistogram(hv), nil
}

func (r *registry) NewHistogramVec(o prometheus.HistogramOpts, labelNames []string) (*prometheus.HistogramVec, error) {
	o.Namespace = r.namespace(o.Namespace)
	o.Subsystem = r.subsystem(o.Subsystem)
	o.ConstLabels = r.mergeConstLabels(o.ConstLabels)

	hv := prometheus.NewHistogramVec(o, labelNames)
	if err := r.Register(hv); err != nil {
		return nil, err
	}

	return hv, nil
}

func (r *registry) NewSummary(o prometheus.SummaryOpts, labelNames []string) (metrics.Histogram, error) {
	sv, err := r.NewSummaryVec(o, labelNames)
	if err != nil {
		return nil, err
	}

	return kitprometheus.NewSummary(sv), nil
}

func (r *registry) NewSummaryVec(o prometheus.SummaryOpts, labelNames []string) (*prometheus.SummaryVec, error) {
	o.Namespace = r.namespace(o.Namespace)
	o.Subsystem = r.subsystem(o.Subsystem)
	o.ConstLabels = r.mergeConstLabels(o.ConstLabels)

	sv := prometheus.NewSummaryVec(o, labelNames)
	if err := r.Register(sv); err != nil {
		return nil, err
	}

	return sv, nil
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
		pco := prometheus.ProcessCollectorOpts{
			Namespace: o.DefaultNamespace,
		}

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
