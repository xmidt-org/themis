package xhttpserver

import (
	"net/http"
	"xhttp"
)

type ResponseHeaders struct {
	Header http.Header
}

func (rh ResponseHeaders) Then(next http.Handler) http.Handler {
	if len(rh.Header) == 0 {
		return next
	}

	header := xhttp.CanonicalizeHeaders(rh.Header)
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		xhttp.AddHeaders(response.Header(), header)
		next.ServeHTTP(response, request)
	})
}

func (rh ResponseHeaders) ThenFunc(next http.HandlerFunc) http.Handler {
	return rh.Then(next)
}
