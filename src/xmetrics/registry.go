package xmetrics

import (
	"github.com/go-kit/kit/metrics"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
)

type Options struct {
	DefaultNamespace        string
	DefaultSubsystem        string
	Pedantic                bool
	DisableGoCollector      bool
	DisableProcessCollector bool
}

type Provider interface {
	prometheus.Registerer
	prometheus.Gatherer

	NewCounter(prometheus.CounterOpts, []string) (metrics.Counter, error)
	NewGauge(prometheus.GaugeOpts, []string) (metrics.Gauge, error)
	NewHistogram(prometheus.HistogramOpts, []string) (metrics.Histogram, error)
	NewSummary(prometheus.SummaryOpts, []string) (metrics.Histogram, error)
}

type provider struct {
	prometheus.Registerer
	prometheus.Gatherer

	defaultNamespace string
	defaultSubsystem string
}

func (p *provider) namespace(v string) string {
	if len(v) > 0 {
		return v
	}

	return p.defaultNamespace
}

func (p *provider) subsystem(v string) string {
	if len(v) > 0 {
		return v
	}

	return p.defaultSubsystem
}

func (p *provider) NewCounter(o prometheus.CounterOpts, labelNames []string) (metrics.Counter, error) {
	o.Namespace = p.namespace(o.Namespace)
	o.Subsystem = p.subsystem(o.Subsystem)

	cv := prometheus.NewCounterVec(o, labelNames)
	if err := p.Register(cv); err != nil {
		return nil, err
	}

	return kitprometheus.NewCounter(cv), nil
}

func (p *provider) NewGauge(o prometheus.GaugeOpts, labelNames []string) (metrics.Gauge, error) {
	o.Namespace = p.namespace(o.Namespace)
	o.Subsystem = p.subsystem(o.Subsystem)

	cv := prometheus.NewGaugeVec(o, labelNames)
	if err := p.Register(cv); err != nil {
		return nil, err
	}

	return kitprometheus.NewGauge(cv), nil
}

func (p *provider) NewHistogram(o prometheus.HistogramOpts, labelNames []string) (metrics.Histogram, error) {
	o.Namespace = p.namespace(o.Namespace)
	o.Subsystem = p.subsystem(o.Subsystem)

	cv := prometheus.NewHistogramVec(o, labelNames)
	if err := p.Register(cv); err != nil {
		return nil, err
	}

	return kitprometheus.NewHistogram(cv), nil
}

func (p *provider) NewSummary(o prometheus.SummaryOpts, labelNames []string) (metrics.Histogram, error) {
	o.Namespace = p.namespace(o.Namespace)
	o.Subsystem = p.subsystem(o.Subsystem)

	cv := prometheus.NewSummaryVec(o, labelNames)
	if err := p.Register(cv); err != nil {
		return nil, err
	}

	return kitprometheus.NewSummary(cv), nil
}

func New(o Options) (Provider, error) {
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

	return &provider{
		Registerer:       pr,
		Gatherer:         pr,
		defaultNamespace: o.DefaultNamespace,
		defaultSubsystem: o.DefaultSubsystem,
	}, nil
}
