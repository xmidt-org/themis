// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package xhttpclient

import (
	"net/http"

	"github.com/xmidt-org/arrange/arrangehttp"
)

// Interface defines the behavior of an HTTP client.  *http.Client implements this interface.
type Interface interface {
	Do(*http.Request) (*http.Response, error)
}

// Options represents the set of configurable options for an HTTP client
type Options struct {
	// HTTPClient is the configuration for the remote claims builder HTTP client.
	HTTPClient arrangehttp.ClientConfig
}

// New fully constructs an http client from a set of options.  NewRoundTripper is used to create the http.RoundTripper.
func New(o Options) (Interface, error) {
	client, err := o.HTTPClient.NewClient()
	if err != nil {
		return nil, err
	}

	return client, o.HTTPClient.Apply(client)
}

// NewCustom uses a set of options and a supplied RoundTripper to create an http client.  Use this function
// when a custom RoundTripper, including decoration, is desired.
func NewCustom(o Options, rt http.RoundTripper) (Interface, error) {
	client, err := o.HTTPClient.NewClient()
	if err != nil {
		return nil, err
	}

	client.Transport = rt

	return client, o.HTTPClient.Apply(client)
}
