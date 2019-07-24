package xmetrics

import (
	"github.com/go-kit/kit/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var empty string

// Labels provides a simple builder for name/value pairs.  Go-kit and prometheus have different
// APIs that use labels.  This type implements a common abstraction for both.
type Labels struct {
	pairs []string
}

// Len returns the number of name/value pairs
func (l *Labels) Len() int {
	if l == nil {
		return 0
	}

	return len(l.pairs) / 2
}

// Reset wipes out the name/value pairs, but does not free the underlying storage
func (l *Labels) Reset() {
	if l != nil {
		for i := 0; i < len(l.pairs); i++ {
			l.pairs[i] = empty
		}

		l.pairs = l.pairs[:0]
	}
}

// Add appends a name/value pair to this Labels instance.  This instance
// is returned for method chaining.
//
// The order in which name/value pairs are added matter.  They should be added in the
// same order as the labels were defined.
func (l *Labels) Add(name, value string) *Labels {
	if l == nil {
		return nil
	}

	l.pairs = append(l.pairs, name, value)
	return l
}

// Labels returns a map of the name/value pairs in this instance.  This method can be
// used with prometheus metrics.
func (l *Labels) Labels() map[string]string {
	if l == nil {
		return nil
	} else if len(l.pairs) > 0 {
		labels := make(map[string]string, len(l.pairs)/2)
		for i := 0; i < len(l.pairs); i += 2 {
			labels[l.pairs[i]] = l.pairs[i+1]
		}

		return labels
	}

	return nil
}

// NamesAndValues returns the name/pair pairs in the order they were added.  This method is useful
// when using go-kit metrics, as the With methods take name/value pairs as a string slice.
func (l *Labels) NamesAndValues() []string {
	if l == nil {
		return nil
	}

	return l.pairs
}

// Values returns a slice of the values only.  This method is useful when using prometheus metrics
// directly, since methods like With take a only the values in the correct order.
func (l *Labels) Values() []string {
	if l == nil {
		return nil
	} else if len(l.pairs) > 0 {
		values := make([]string, len(l.pairs)/2)
		for i := 1; i < len(l.pairs); i += 2 {
			values[i>>1] = l.pairs[i]
		}

		return values
	}

	return nil
}

type Adder interface {
	Add(*Labels, float64)
}

type LabelledCounter struct {
	metrics.Counter
}

func (lc LabelledCounter) Add(l *Labels, v float64) {
	lc.Counter.With(l.NamesAndValues()...).Add(v)
}

type LabelledCounterVec struct {
	*prometheus.CounterVec
}

func (lcv LabelledCounterVec) Add(l *Labels, v float64) {
	lcv.CounterVec.WithLabelValues(l.Values()...).Add(v)
}

type Setter interface {
	Set(*Labels, float64)
}

type AdderSetter interface {
	Adder
	Setter
}

type LabelledGauge struct {
	metrics.Gauge
}

func (lg LabelledGauge) Add(l *Labels, v float64) {
	lg.Gauge.With(l.NamesAndValues()...).Add(v)
}

func (lg LabelledGauge) Set(l *Labels, v float64) {
	lg.Gauge.With(l.NamesAndValues()...).Set(v)
}

type LabelledGaugeVec struct {
	*prometheus.GaugeVec
}

func (lgv LabelledGaugeVec) Add(l *Labels, v float64) {
	lgv.GaugeVec.WithLabelValues(l.Values()...).Add(v)
}

func (lgv LabelledGaugeVec) Set(l *Labels, v float64) {
	lgv.GaugeVec.WithLabelValues(l.Values()...).Set(v)
}

type Observer interface {
	Observe(*Labels, float64)
}

type LabelledHistogram struct {
	metrics.Histogram
}

func (lh LabelledHistogram) Observe(l *Labels, v float64) {
	lh.Histogram.With(l.NamesAndValues()...).Observe(v)
}

type LabelledObserverVec struct {
	prometheus.ObserverVec
}

func (lov LabelledObserverVec) Observe(l *Labels, v float64) {
	lov.ObserverVec.WithLabelValues(l.Values()...).Observe(v)
}
