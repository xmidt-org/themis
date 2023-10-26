// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package xmetrics

import (
	"bytes"

	"github.com/go-kit/kit/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var empty string

// Labels provides a simple builder for name/value pairs.  Go-kit and prometheus have different
// APIs that use labels.  This type implements a common abstraction for both.
//
// A nil Labels is valid, and behaves exactly like and empty Labels would.
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

func (l *Labels) String() string {
	if l == nil || len(l.pairs) == 0 {
		return empty
	}

	var output bytes.Buffer
	for i := 0; i < len(l.pairs); i += 2 {
		if i > 0 {
			output.WriteRune(',')
		}

		output.WriteString(l.pairs[i])
		output.WriteRune('=')
		output.WriteString(l.pairs[i+1])
	}

	return output.String()
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

// Adder is a strategy for adding a delta to a metric, with optional labels applied
type Adder interface {
	// Add increments the underlying metric, applying Labels if non-nil and non-empty.
	// Note that this method should only be used with positive values.
	Add(*Labels, float64)
}

// LabelledCounter is an Adder which uses a go-kit Counter
type LabelledCounter struct {
	metrics.Counter
}

func (lc LabelledCounter) Add(l *Labels, v float64) {
	lc.Counter.With(l.NamesAndValues()...).Add(v)
}

// LabelledCounterVec is an Adder which uses a prometheus CounterVec
type LabelledCounterVec struct {
	*prometheus.CounterVec
}

func (lcv LabelledCounterVec) Add(l *Labels, v float64) {
	lcv.CounterVec.WithLabelValues(l.Values()...).Add(v)
}

// Setter is a strategy for setting values on a metric, with optional labels applied
type Setter interface {
	// Set puts a value to the underlying metric, applying Labels if non-nil and non-empty
	Set(*Labels, float64)
}

// GaugeAdder is like Adder, but specific to gauges.  Client code can consume this interface
// to prevent counters from being used where a gauge is specifically needed.  With most metrics
// backends, counters can only have positive values added while gauges allow adding any value.
// Use of this interface allows the compiler to prevent misconfiguration.
type GaugeAdder interface {
	// GaugeAdd adds a delta to the underlying metric, applying Labels if non-nil and non-empty.
	// This method can be used with any value, not just positive values.
	GaugeAdd(*Labels, float64)
}

// LabelledGauge provides Adder, Setter, and GaugeAdder support for a go-kit Gauge
type LabelledGauge struct {
	metrics.Gauge
}

func (lg LabelledGauge) Add(l *Labels, v float64) {
	lg.Gauge.With(l.NamesAndValues()...).Add(v)
}

func (lg LabelledGauge) Set(l *Labels, v float64) {
	lg.Gauge.With(l.NamesAndValues()...).Set(v)
}

func (lg LabelledGauge) GaugeAdd(l *Labels, v float64) {
	lg.Gauge.With(l.NamesAndValues()...).Add(v)
}

// LabelledGaugeVec provides Adder, Setter, and GaugeAdder support for a prometheus GaugeVec
type LabelledGaugeVec struct {
	*prometheus.GaugeVec
}

func (lgv LabelledGaugeVec) Add(l *Labels, v float64) {
	lgv.GaugeVec.WithLabelValues(l.Values()...).Add(v)
}

func (lgv LabelledGaugeVec) Set(l *Labels, v float64) {
	lgv.GaugeVec.WithLabelValues(l.Values()...).Set(v)
}

func (lgv LabelledGaugeVec) GaugeAdd(l *Labels, v float64) {
	lgv.GaugeVec.WithLabelValues(l.Values()...).Add(v)
}

// Observer is a strategy for observing series of values
type Observer interface {
	// Observe posts a value to the underlying metric, applying Labels if non-nil and non-empty
	Observe(*Labels, float64)
}

// LabelledHistogram is an Observer backed by a go-kit Histogram
type LabelledHistogram struct {
	metrics.Histogram
}

func (lh LabelledHistogram) Observe(l *Labels, v float64) {
	lh.Histogram.With(l.NamesAndValues()...).Observe(v)
}

// LabelledObserverVec is an Observer backed by a prometheus ObserverVec, which can be either
// a HistogramVec or a SummaryVec
type LabelledObserverVec struct {
	prometheus.ObserverVec
}

func (lov LabelledObserverVec) Observe(l *Labels, v float64) {
	lov.ObserverVec.WithLabelValues(l.Values()...).Observe(v)
}
