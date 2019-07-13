package xloghttp

import (
	"net/http"
	"strings"
	"xlog"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

type ParameterBuilder func(*http.Request, []interface{}) []interface{}

func Method(original *http.Request, parameters []interface{}) []interface{} {
	return append(
		parameters,
		"requestMethod",
		original.Method,
	)
}

func URI(original *http.Request, parameters []interface{}) []interface{} {
	return append(
		parameters,
		"requestURI",
		original.RequestURI,
	)
}

func RemoteAddress(original *http.Request, parameters []interface{}) []interface{} {
	return append(
		parameters,
		"remoteAddr",
		original.RemoteAddr,
	)
}

func Header(name string) ParameterBuilder {
	return func(original *http.Request, parameters []interface{}) []interface{} {
		value := original.Header.Get(name)
		return append(
			parameters,
			name,
			value,
		)
	}
}

func Parameter(name string) ParameterBuilder {
	return func(original *http.Request, parameters []interface{}) []interface{} {
		values := original.Form[name]
		return append(
			parameters,
			name,
			strings.Join(values, ","),
		)
	}
}

func Variable(name string) ParameterBuilder {
	return func(original *http.Request, parameters []interface{}) []interface{} {
		value := mux.Vars(original)[name]
		return append(
			parameters,
			name,
			value,
		)
	}
}

func WithRequest(original *http.Request, l log.Logger, b ...ParameterBuilder) *http.Request {
	if len(b) > 0 {
		var parameters []interface{}
		for _, f := range b {
			parameters = f(original, parameters)
		}

		l = log.WithPrefix(
			l,
			parameters...,
		)
	}

	return original.WithContext(
		xlog.With(
			original.Context(),
			l,
		),
	)
}

type Logging struct {
	Base     log.Logger
	Builders []ParameterBuilder
}

func (l Logging) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		next.ServeHTTP(
			response,
			WithRequest(request, l.Base, l.Builders...),
		)
	})
}
