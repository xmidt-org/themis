package xmetricshttp

import (
	"net/http"
	"strconv"
	"xmetrics"
)

const (
	// CodeLabel is the standard metric label name used in this package for the response code
	CodeLabel = "code"

	// MethodLabel is the standard metric label name used in this package for the request method
	MethodLabel = "method"
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

// StandardLabeller produces code and method name/value pairs.  To use this type as a ServerLabeller, the http.ResponseWriter
// must implement StatusCoder.
type StandardLabeller struct{}

// ServerLabels appends the standard labels, code and method, for this package.  The http.ResponseWriter must
// implement StatusCoder, or this method will panic.  Decoration of the http.ResponseWriter is left to other packages.
func (sl StandardLabeller) ServerLabels(response http.ResponseWriter, request *http.Request, l *xmetrics.Labels) {
	l.Add(CodeLabel, strconv.Itoa(response.(StatusCoder).StatusCode()))
	l.Add(MethodLabel, request.Method)
}

func (sl StandardLabeller) ClientLabels(response *http.Response, request *http.Request, l *xmetrics.Labels) {
	l.Add(CodeLabel, strconv.Itoa(response.StatusCode))
	l.Add(MethodLabel, request.Method)
}
