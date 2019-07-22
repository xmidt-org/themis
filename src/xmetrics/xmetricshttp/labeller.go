package xmetricshttp

import (
	"net/http"
	"strconv"
	"xmetrics"
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

// ServerLabeller is a strategy for producing metrics label/value pairs from a serverside HTTP request
type ServerLabeller interface {
	ServerLabels(http.ResponseWriter, *http.Request, *xmetrics.Labels)
}

type ServerLabellerFunc func(http.ResponseWriter, *http.Request, *xmetrics.Labels)

func (slf ServerLabellerFunc) ServerLabels(response http.ResponseWriter, request *http.Request, l *xmetrics.Labels) {
	slf(response, request, l)
}

// ServerLabellers is a set of ServerLabeller strategies to be executed in sequence
type ServerLabellers []ServerLabeller

func (sl ServerLabellers) ServerLabels(response http.ResponseWriter, request *http.Request, l *xmetrics.Labels) {
	for _, slf := range sl {
		slf.ServerLabels(response, request, l)
	}
}

// ClientLabeller is a strategy for producing metrics label/value pairs from a clientside HTTP request
type ClientLabeller interface {
	ClientLabels(*http.Response, *http.Request, *xmetrics.Labels)
}

type ClientLabellerFunc func(*http.Response, *http.Request, *xmetrics.Labels)

func (clf ClientLabellerFunc) ClientLabels(response *http.Response, request *http.Request, l *xmetrics.Labels) {
	clf(response, request, l)
}

// ClientLabellers is a set of ClientLabeller strategies to be executed in sequence
type ClientLabellers []ClientLabeller

func (cl ClientLabellers) ServerLabels(response *http.Response, request *http.Request, l *xmetrics.Labels) {
	for _, clf := range cl {
		clf.ClientLabels(response, request, l)
	}
}

// EmptyLabeller is both a server and client labeller which produces no labels
type EmptyLabeller struct{}

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

func (cl CodeLabeller) ServerLabels(response http.ResponseWriter, _ *http.Request, l *xmetrics.Labels) {
	name := cl.Name
	if len(name) == 0 {
		name = DefaultCodeLabel
	}

	l.Add(name, strconv.Itoa(response.(StatusCoder).StatusCode()))
}

func (cl CodeLabeller) ClientLabels(response *http.Response, _ *http.Request, l *xmetrics.Labels) {
	name := cl.Name
	if len(name) == 0 {
		name = DefaultCodeLabel
	}

	l.Add(name, strconv.Itoa(response.StatusCode))
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

// MethodLabeller provides both server and client labelling for the HTTP request method.
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

func (ml MethodLabeller) labels(request *http.Request, l *xmetrics.Labels) {
	name := ml.Name
	if len(name) == 0 {
		name = DefaultMethodLabel
	}

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

	l.Add(name, value)
}

func (ml MethodLabeller) ServerLabels(_ http.ResponseWriter, request *http.Request, l *xmetrics.Labels) {
	ml.labels(request, l)
}

func (ml MethodLabeller) ClientLabels(_ *http.Response, request *http.Request, l *xmetrics.Labels) {
	ml.labels(request, l)
}
