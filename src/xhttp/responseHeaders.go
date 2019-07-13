package xhttp

import (
	"net/http"

	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type ResponseHeaders func(http.Handler) http.Handler

// NewResponseHeaders produces an Alice-style constructor that sets static response headers on
// every response.
func NewResponseHeaders(h http.Header) ResponseHeaders {
	if len(h) == 0 {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			for name, values := range h {
				response.Header()[name] = values
			}

			next.ServeHTTP(response, request)
		})
	}
}

type ResponseHeadersIn struct {
	fx.In

	Viper *viper.Viper
}

func ProvideResponseHeaders(configKey string) func(ResponseHeadersIn) (ResponseHeaders, error) {
	return func(in ResponseHeadersIn) (ResponseHeaders, error) {
		var o map[string]string
		if err := in.Viper.UnmarshalKey(configKey, &o); err != nil {
			return nil, err
		}

		h := make(http.Header, len(o))
		for k, v := range o {
			h.Set(k, v)
		}

		return NewResponseHeaders(h), nil
	}
}
