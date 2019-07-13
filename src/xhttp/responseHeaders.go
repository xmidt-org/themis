package xhttp

import (
	"net/http"

	"github.com/justinas/alice"
)

// ResponseHeaders produces an Alice-style constructor that sets static response headers on
// every response.
func ResponseHeaders(h http.Header) alice.Constructor {
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
