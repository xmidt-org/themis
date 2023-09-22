// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package xhttpserver

import (
	"net/http"
	"strconv"
)

// Constant describes the available options for creating a ConstantHandler
type Constant struct {
	// StatusCode is the HTTP response code returned from all transactions.  If unset, http.StatusOK is assumed.
	StatusCode int

	// Header describes any response HTTP headers.  If a Body is supplied, the Content-Length header is automatically inserted.
	// If a Content-Type header is desired, use this field.
	Header http.Header

	// Body is the constant body returned by all HTTP responses.  If unset, no body is written.  Note that the
	// Content-Length header is always written.  If no Body is set, Content-Length will be zero.
	Body []byte
}

// NewHandler produces a ConstantHandler from a set of options.  Headers are preprocessed and canoncilized.
// Each invocation of this method produces a distinct ConstantHandler that is unaffected by changes to the
// defining Constant instance.
func (c Constant) NewHandler() *ConstantHandler {
	ch := &ConstantHandler{
		header: make(http.Header, len(c.Header)),
		body:   c.Body,
	}

	if c.StatusCode > 0 {
		ch.statusCode = c.StatusCode
	} else {
		ch.statusCode = http.StatusOK
	}

	if len(c.Header) > 0 {
		for name, values := range c.Header {
			for _, value := range values {
				ch.header.Add(name, value)
			}
		}
	}

	ch.header.Set("Content-Length", strconv.Itoa(len(c.Body)))
	return ch
}

// ConstantHandler is an http.Handler that produces constant content described by a ConstantOptions
type ConstantHandler struct {
	statusCode int
	header     http.Header
	body       []byte
}

func (ch *ConstantHandler) ServeHTTP(response http.ResponseWriter, _ *http.Request) {
	for name, values := range ch.header {
		response.Header()[name] = values
	}

	response.WriteHeader(ch.statusCode)
	if len(ch.body) > 0 {
		response.Write(ch.body)
	}
}
