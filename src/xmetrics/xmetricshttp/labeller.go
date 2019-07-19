package xmetricshttp

import (
	"net/http"
	"strconv"
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
	ServerLabels(http.ResponseWriter, *http.Request, []string) []string
}

type ServerLabellerFunc func(http.ResponseWriter, *http.Request, []string) []string

func (slf ServerLabellerFunc) ServerLabels(response http.ResponseWriter, request *http.Request, lvs []string) []string {
	return slf(response, request, lvs)
}

// ServerLabellers is a set of ServerLabeller strategies to be executed in sequence
type ServerLabellers []ServerLabeller

func (sl ServerLabellers) ServerLabels(response http.ResponseWriter, request *http.Request, lvs []string) []string {
	for _, slf := range sl {
		lvs = slf.ServerLabels(response, request, lvs)
	}

	return lvs
}

// ClientLabeller is a strategy for producing metrics label/value pairs from a clientside HTTP request
type ClientLabeller interface {
	ClientLabels(*http.Response, *http.Request, []string) []string
}

type ClientLabellerFunc func(*http.Response, *http.Request, []string) []string

func (clf ClientLabellerFunc) ClientLabels(response *http.Response, request *http.Request, lvs []string) []string {
	return clf(response, request, lvs)
}

// ClientLabellers is a set of ClientLabeller strategies to be executed in sequence
type ClientLabellers []ClientLabeller

func (cl ClientLabellers) ServerLabels(response *http.Response, request *http.Request, lvs []string) []string {
	for _, clf := range cl {
		lvs = clf.ClientLabels(response, request, lvs)
	}

	return lvs
}

// EmptyLabeller is both a server and client labeller which produces no labels
type EmptyLabeller struct{}

func (el EmptyLabeller) ServerLabels(_ http.ResponseWriter, _ *http.Request, lvs []string) []string {
	return lvs
}

func (el EmptyLabeller) ClientLabels(_ *http.Response, _ *http.Request, lvs []string) []string {
	return lvs
}

// StandardLabeller produces code and method name/value pairs.  To use this type as a ServerLabeller, the http.ResponseWriter
// must implement StatusCoder.
type StandardLabeller struct{}

// ServerLabels appends the standard labels, code and method, for this package.  The http.ResponseWriter must
// implement StatusCoder, or this method will panic.  Decoration of the http.ResponseWriter is left to other packages.
func (sl StandardLabeller) ServerLabels(response http.ResponseWriter, request *http.Request, lvs []string) []string {
	return append(lvs,
		CodeLabel,
		strconv.Itoa(response.(StatusCoder).StatusCode()),
		MethodLabel,
		request.Method,
	)
}

func (sl StandardLabeller) ClientLabels(response *http.Response, request *http.Request, lvs []string) []string {
	return append(lvs,
		CodeLabel,
		strconv.Itoa(response.StatusCode),
		MethodLabel,
		request.Method,
	)
}
