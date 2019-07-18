package xmetricshttp

import (
	"net/http"

	"github.com/go-kit/kit/metrics"
)

// Reporter is the strategy interface used for reporting HTTP-related metrics.  Implementations
// are responsible for interacting with metrics infrastructure to record the value.
type Reporter interface {
	// Report reports a value.  The float64 value is interpreted differently based on the calling code.
	// It can be a count, a duration, an observed value, etc.
	Report(http.ResponseWriter, *http.Request, float64)
}

type ReporterFunc func(http.ResponseWriter, *http.Request, float64)

func (rf ReporterFunc) Report(response http.ResponseWriter, request *http.Request, value float64) {
	rf(response, request, value)
}

// DiscardReporter is a Reporter that simply discards any values
type DiscardReporter struct{}

func (dr DiscardReporter) Report(http.ResponseWriter, *http.Request, float64) {
}

// CounterReporter is a reporter that interprets values as deltas to add to a go-kit metrics.Counter
type CounterReporter struct {
	// Metric is the required go-kit Counter that receives observations from this reporter
	Metric metrics.Counter

	// Labeller is the required strategy for producing name/value pairs from a serverside HTTP request.
	// If no labels are used, set this field to EmptyLabeller{}.
	Labeller Labeller
}

func (cr CounterReporter) Report(response http.ResponseWriter, request *http.Request, delta float64) {
	var (
		c           = cr.Metric
		labelValues = cr.Labeller.LabelValuesFor(response, request)
	)

	if len(labelValues) > 0 {
		c = c.With(labelValues...)
	}

	c.Add(delta)
}

// GaugeReporter is a reporter that interprets values according to the Add member: either as a delta to add similar
// to a counter or a value to set.
type GaugeReporter struct {
	// Metric is the required go-kit Gauge that receives observations from this reporter
	Metric metrics.Gauge

	// Labeller is the required strategy for producing name/value pairs from a serverside HTTP request.
	// If no labels are used, set this field to EmptyLabeller{}.
	Labeller Labeller

	// Add indicates how to interpret values.  If true, Metric.Add is used.  If false (the default), metric.Set is used.
	Add bool
}

func (gr GaugeReporter) Report(response http.ResponseWriter, request *http.Request, value float64) {
	var (
		c           = gr.Metric
		labelValues = gr.Labeller.LabelValuesFor(response, request)
	)

	if len(labelValues) > 0 {
		c = c.With(labelValues...)
	}

	if gr.Add {
		c.Add(value)
	} else {
		c.Set(value)
	}
}

// HistogramReporter is a reporter that interprets values as observations on a go-kit metrics.Histogram
type HistogramReporter struct {
	// Metric is the required go-kit Histogram that receives observations from this reporter
	Metric metrics.Histogram

	// Labeller is the required strategy for producing name/value pairs from a serverside HTTP request.
	// If no labels are used, set this field to EmptyLabeller{}.
	Labeller Labeller
}

func (hr HistogramReporter) Report(response http.ResponseWriter, request *http.Request, value float64) {
	var (
		c           = hr.Metric
		labelValues = hr.Labeller.LabelValuesFor(response, request)
	)

	if len(labelValues) > 0 {
		c = c.With(labelValues...)
	}

	c.Observe(value)
}
