package xhttpclient

import "net/http"

// RequestHeaders provides a RoundTripper constructor that inserts a constant set of headers
// into each request
type RequestHeaders struct {
	Header http.Header
}

func (rh RequestHeaders) Then(next http.RoundTripper) http.RoundTripper {
	if len(rh.Header) == 0 {
		return next
	}

	return RoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		for name, values := range rh.Header {
			for _, value := range values {
				request.Header.Add(name, value)
			}
		}

		return next.RoundTrip(request)
	})
}

func (rh RequestHeaders) ThenFunc(next RoundTripperFunc) http.RoundTripper {
	return rh.Then(next)
}
