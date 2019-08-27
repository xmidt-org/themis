package xloghttp

import (
	"net/http"
	"strings"

	"github.com/xmidt-org/themis/xlog"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

// Parameters is a simple builder for logging key/value pairs
type Parameters struct {
	values []interface{}
}

func (p *Parameters) Add(key, value interface{}) *Parameters {
	p.values = append(p.values, key, value)
	return p
}

func (p *Parameters) Use(base log.Logger) log.Logger {
	if len(p.values) > 0 {
		return log.With(base, p.values...)
	}

	return base
}

// ParameterBuilder appends logging key/value pairs to be used with a contextual request logger
type ParameterBuilder func(*http.Request, *Parameters)

// ParameterBuilders is a simple slice type for ParameterBuilder instances
type ParameterBuilders []ParameterBuilder

// Method returns a ParameterBuilder that adds the HTTP request method as a logging key/value pair
func Method(key string) ParameterBuilder {
	return func(original *http.Request, p *Parameters) {
		p.Add(key, original.Method)
	}
}

// URI returns a ParameterBuilder that adds the HTTP request URI as a logging key/value pair.
// Note that only the path is emitted as a logging value.  The query string and any host:port information
// is omitted.
func URI(key string) ParameterBuilder {
	return func(original *http.Request, p *Parameters) {
		p.Add(key, original.URL.Path)
	}
}

// RemoteAddress is a ParameterBuilder that adds the HTTP remote address as a logging key/value pair
func RemoteAddress(key string) ParameterBuilder {
	return func(original *http.Request, p *Parameters) {
		p.Add(key, original.RemoteAddr)
	}
}

// Header returns a ParameterBuilder that appends the given HTTP header as a key/value pair
func Header(name string) ParameterBuilder {
	name = http.CanonicalHeaderKey(name)
	return func(original *http.Request, p *Parameters) {
		p.Add(name, strings.Join(original.Header[name], ","))
	}
}

// Parameter returns a ParameterBuilder that appends the given HTTP query or form parameter as a key/value pair
func Parameter(name string) ParameterBuilder {
	return func(original *http.Request, p *Parameters) {
		p.Add(name, strings.Join(original.Form[name], ","))
	}
}

// Variable returns a ParameterBuilder that appends the given gorilla/mux path variable as a key/value pair
func Variable(name string) ParameterBuilder {
	return func(original *http.Request, p *Parameters) {
		p.Add(name, mux.Vars(original)[name])
	}
}

// WithRequest produces a new http.Request with a contextual logger bound to the context.
func WithRequest(original *http.Request, l log.Logger, b ...ParameterBuilder) *http.Request {
	if len(b) > 0 {
		var p Parameters
		for _, f := range b {
			f(original, &p)
		}

		l = p.Use(l)
	}

	return original.WithContext(
		xlog.With(
			original.Context(),
			l,
		),
	)
}

// Logging provides an Alice-style decorator that attaches a contextual logger to requests
type Logging struct {
	Base     log.Logger
	Builders ParameterBuilders
}

func (l Logging) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		next.ServeHTTP(
			response,
			WithRequest(request, l.Base, l.Builders...),
		)
	})
}

func (l Logging) ThenFunc(next http.HandlerFunc) http.Handler {
	return l.Then(next)
}
