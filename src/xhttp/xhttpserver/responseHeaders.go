package xhttpserver

import (
	"net/http"

	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type ResponseHeaders struct {
	Headers http.Header
}

func (rh ResponseHeaders) Then(next http.Handler) http.Handler {
	if len(rh.Headers) == 0 {
		return next
	}

	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		for name, values := range rh.Headers {
			for _, value := range values {
				response.Header().Add(name, value)
			}
		}

		next.ServeHTTP(response, request)
	})
}

type ResponseHeadersIn struct {
	fx.In

	Viper *viper.Viper
}

func UnmarshalResponseHeaders(configKey string) func(ResponseHeadersIn) (ResponseHeaders, error) {
	return func(in ResponseHeadersIn) (ResponseHeaders, error) {
		var o map[string]string
		if err := in.Viper.UnmarshalKey(configKey, &o); err != nil {
			return ResponseHeaders{}, err
		}

		if len(o) == 0 {
			return ResponseHeaders{}, nil
		}

		h := make(http.Header, len(o))
		for k, v := range o {
			h.Set(k, v)
		}

		return ResponseHeaders{Headers: h}, nil
	}
}
