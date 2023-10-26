// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package xmetricshttp

import (
	"net/http"
	"strconv"

	"github.com/xmidt-org/themis/xmetrics"
)

const (
	DefaultCodeLabel   = "code"
	DefaultMethodLabel = "method"
	DefaultOther       = "other"
)

// StatusCoder is expected to be implemented by http.ResponseWriters that participate in metrics.
// Decorating the http.ResponseWriter is left to other packages.
type StatusCoder interface {
	StatusCode() int
}

// LabelNames is a strategy for determining the names for metrics labels
type LabelNames interface {
	// LabelNames returns the ordered set of metrics label names.  The order in which these
	// names occur in the slice will match any labeller strategies in a ClientLabeller or a ServerLabeller.
	LabelNames() []string
}

// ServerLabeller is a strategy for producing metrics label/value pairs from a serverside HTTP request
type ServerLabeller interface {
	LabelNames

	// ServerLabels applies the labels this strategy provides.  The order the labels are applies
	// must match the order of names returned by LabelNames.
	ServerLabels(http.ResponseWriter, *http.Request, *xmetrics.Labels)
}

// ServerLabellers is a set of ServerLabeller strategies to be executed in sequence.  Keys are label names.
// This type guarantees a consistent ordering for both the labels and labellers.
//
// A default ServerLabellers is ready to use and can be built up using Add.  A nil ServerLabellers is also
// valid, and acts as a no-op.
type ServerLabellers struct {
	labelNames []string
	labellers  []ServerLabeller
}

func NewServerLabellers(labellers ...ServerLabeller) *ServerLabellers {
	sl := &ServerLabellers{
		labelNames: make([]string, len(labellers)), // just an optimization step
		labellers:  append([]ServerLabeller{}, labellers...),
	}

	for _, l := range labellers {
		sl.labelNames = append(sl.labelNames, l.LabelNames()...)
	}

	return sl
}

func (sl *ServerLabellers) LabelNames() []string {
	if sl == nil {
		return nil
	}

	return sl.labelNames
}

func (sl *ServerLabellers) ServerLabels(response http.ResponseWriter, request *http.Request, l *xmetrics.Labels) {
	if sl != nil {
		for _, labeller := range sl.labellers {
			labeller.ServerLabels(response, request, l)
		}
	}
}

// ClientLabeller is a strategy for producing metrics label/value pairs from a clientside HTTP request
type ClientLabeller interface {
	LabelNames

	// ClientLabels applies the labels this strategy provides.  The order the labels are applies
	// must match the order of names returned by LabelNames.
	ClientLabels(*http.Response, *http.Request, *xmetrics.Labels)
}

// ClientLabellers is a set of ClientLabeller strategies to be executed in sequence.  Keys are label names.
// This type guarantees a consistent ordering for both the labels and labellers.
//
// A default ClientLabellers is ready to use and can be built up using Add.  A nil ClientLabellers is valid
// and acts as a no-op.
type ClientLabellers struct {
	labelNames []string
	labellers  []ClientLabeller
}

func NewClientLabellers(labellers ...ClientLabeller) *ClientLabellers {
	cl := &ClientLabellers{
		labelNames: make([]string, len(labellers)), // just an optimization step
		labellers:  append([]ClientLabeller{}, labellers...),
	}

	for _, l := range labellers {
		cl.labelNames = append(cl.labelNames, l.LabelNames()...)
	}

	return cl
}

// LabelNames returns the label names in the order they were added.  This is the same order that
// ClientLabels applies labels to a metric.
func (cl *ClientLabellers) LabelNames() []string {
	if cl == nil {
		return nil
	}

	return cl.labelNames
}

func (cl *ClientLabellers) ClientLabels(response *http.Response, request *http.Request, l *xmetrics.Labels) {
	if cl != nil {
		for _, labeller := range cl.labellers {
			labeller.ClientLabels(response, request, l)
		}
	}
}

// EmptyLabeller is both a server and client labeller which produces no labels
type EmptyLabeller struct{}

func (el EmptyLabeller) LabelNames() []string {
	return nil
}

func (el EmptyLabeller) ServerLabels(http.ResponseWriter, *http.Request, *xmetrics.Labels) {
}

func (el EmptyLabeller) ClientLabels(*http.Response, *http.Request, *xmetrics.Labels) {
}

// CodeLabeller provides both ServerLabeller and ClientLabeller functionality for HTTP response codes.
// For servers, the http.ResponseWriter must implement the StatusCode interface, or this labeller will panic.
type CodeLabeller struct {
	// Name is the name of the label to apply.  If unset, DefaultCodeLabel is used.
	Name string
}

func (cl CodeLabeller) name() string {
	if len(cl.Name) > 0 {
		return cl.Name
	}

	return DefaultCodeLabel
}

func (cl CodeLabeller) LabelNames() []string {
	return []string{cl.name()}
}

func (cl CodeLabeller) ServerLabels(response http.ResponseWriter, _ *http.Request, l *xmetrics.Labels) {
	l.Add(cl.name(), strconv.Itoa(response.(StatusCoder).StatusCode()))
}

func (cl CodeLabeller) ClientLabels(response *http.Response, _ *http.Request, l *xmetrics.Labels) {
	l.Add(cl.name(), strconv.Itoa(response.StatusCode))
}

var defaultTrackedMethods = map[string]bool{
	"GET":     true,
	"HEAD":    true,
	"POST":    true,
	"PUT":     true,
	"DELETE":  true,
	"TRACE":   true,
	"CONNECT": true,
}

// MethodLabeller provides both server and client labeling for the HTTP request method.
type MethodLabeller struct {
	// Name is the name of the label to apply.  If unset, DefaultMethodLabel is used.
	Name string

	// TrackedMethods is a set of HTTP methods that are tracked by this labeller.  Methods that
	// don't have keys in this map use the Other value instead.  If unset, a default map
	// is used that includes all standard HTTP methods.
	TrackedMethods map[string]bool

	// Other is the value used for methods that do not have a key in the TrackedMethods map.
	// If unset, DefaultOther is used.
	Other string
}

func (ml MethodLabeller) name() string {
	name := ml.Name
	if len(name) == 0 {
		name = DefaultMethodLabel
	}

	return name
}

func (ml MethodLabeller) labels(request *http.Request, l *xmetrics.Labels) {
	value := request.Method
	if len(ml.TrackedMethods) > 0 {
		if !ml.TrackedMethods[value] {
			value = ml.Other
		}
	} else if !defaultTrackedMethods[value] {
		value = ml.Other
	}

	if len(value) == 0 {
		value = DefaultOther
	}

	l.Add(ml.name(), value)
}

func (ml MethodLabeller) LabelNames() []string {
	return []string{ml.name()}
}

func (ml MethodLabeller) ServerLabels(_ http.ResponseWriter, request *http.Request, l *xmetrics.Labels) {
	ml.labels(request, l)
}

func (ml MethodLabeller) ClientLabels(_ *http.Response, request *http.Request, l *xmetrics.Labels) {
	ml.labels(request, l)
}
