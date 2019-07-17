package xhttpserver

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

var (
	// ErrHijackerNotSupported is returned by TrackingWriter.Hijack when the underlying
	// http.ResponseWriter does not implement http.Hijacker.
	ErrHijackerNotSupported = errors.New("http.Hijacker is not supported")
)

// TrackingWriter is a decorated http.ResponseWriter that allows visibility into
// various items written to the response.
//
// Implementations always implement the optional interfaces.  In cases where the underlying
// http.ResponseWriter does not implement that interface, e.g. http.Pusher, either an error
// is returned or the method is a no-op.
type TrackingWriter interface {
	http.ResponseWriter
	http.Hijacker
	http.Pusher
	http.Flusher

	// Hijacked returns true if the underlying network connection has been hijacked
	Hijacked() bool

	// StatusCode returns the HTTP response code that was written using WriteHeader.
	// If WriteHeader was not called, http.StatusOK is returned.
	StatusCode() int

	// BytesWritten returns the total bytes written to the response body via Write.
	BytesWritten() int
}

// NewTrackingWriter decorates an existing response writer and allows visibility
// into certain items written to the response.
func NewTrackingWriter(next http.ResponseWriter) TrackingWriter {
	delegate := &trackingWriter{
		next: next,
	}

	return delegate
}

type trackingWriter struct {
	next http.ResponseWriter

	hijacked     bool
	statusCode   int
	bytesWritten int
}

func (dw *trackingWriter) Hijacked() bool {
	return dw.hijacked
}

func (dw *trackingWriter) StatusCode() int {
	if dw.statusCode > 0 {
		return dw.statusCode
	}

	return http.StatusOK
}

func (dw *trackingWriter) BytesWritten() int {
	return dw.bytesWritten
}

func (dw *trackingWriter) Header() http.Header {
	return dw.next.Header()
}

func (dw *trackingWriter) Write(b []byte) (int, error) {
	c, err := dw.next.Write(b)
	if c > 0 {
		dw.bytesWritten += c
	}

	return c, err
}

func (dw *trackingWriter) WriteHeader(statusCode int) {
	if dw.statusCode <= 0 {
		dw.statusCode = statusCode
	}

	dw.next.WriteHeader(statusCode)
}

func (dw *trackingWriter) Flush() {
	if f, ok := dw.next.(http.Flusher); ok {
		f.Flush()
	}
}

func (dw *trackingWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := dw.next.(http.Hijacker); ok {
		c, rw, err := h.Hijack()
		if err == nil {
			dw.hijacked = true
		}

		return c, rw, err
	}

	return nil, nil, ErrHijackerNotSupported
}

func (dw *trackingWriter) Push(target string, opts *http.PushOptions) error {
	if h, ok := dw.next.(http.Pusher); ok {
		return h.Push(target, opts)
	}

	return http.ErrNotSupported
}

// TrackWriter is an Alice-style constructor that wraps the response writer as a TrackingWriter
func TrackWriter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(original http.ResponseWriter, request *http.Request) {
		next.ServeHTTP(
			NewTrackingWriter(original),
			request,
		)
	})
}
