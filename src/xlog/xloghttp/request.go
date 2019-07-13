package xloghttp

import (
	"net/http"
	"strings"
	"xlog"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
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

func WithRequest(original *http.Request, l log.Logger, b []ParameterBuilder) *http.Request {
	var parameters []interface{}
	for _, f := range b {
		parameters = f(original, parameters)
	}

	ctx := xlog.With(
		original.Context(),
		log.WithPrefix(
			l,
			parameters...,
		),
	)

	return original.WithContext(ctx)
}

type Constructor alice.Constructor

func NewConstructor(l log.Logger, b ...ParameterBuilder) Constructor {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			next.ServeHTTP(
				response,
				WithRequest(request, l, b),
			)
		})
	}
}
