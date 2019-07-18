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

// Labeller is a strategy for producing metrics label/value pairs from a serverside HTTP request
type Labeller interface {
	LabelValuesFor(http.ResponseWriter, *http.Request) []string
}

type LabellerFunc func(http.ResponseWriter, *http.Request) []string

func (lf LabellerFunc) LabelValuesFor(response http.ResponseWriter, request *http.Request) []string {
	return lf(response, request)
}

// EmptyLabeller is a Labeller which produces no label name/value pairs
type EmptyLabeller struct{}

func (el EmptyLabeller) LabelValuesFor(http.ResponseWriter, *http.Request) []string {
	return nil
}

// StandardLabeller produces code and method name/value pairs.  To use this Labeller, the http.ResponseWriter
// must implement StatusCoder.  Otherwise, this Labeller will panic.
type StandardLabeller struct{}

func (sl StandardLabeller) LabelValuesFor(response http.ResponseWriter, request *http.Request) []string {
	return []string{
		CodeLabel,
		strconv.Itoa(response.(StatusCoder).StatusCode()),
		MethodLabel,
		request.Method,
	}
}
