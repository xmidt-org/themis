// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package token

import (
	"context"
	"net/http"
	"time"

	"github.com/go-kit/kit/endpoint"
)

const (
	JWTExpHeader = "Expires"
)

// NewIssueEndpoint returns a go-kit endpoint for a token factory's NewToken method
func NewIssueEndpoint(f Factory, headerClaims map[string]string) endpoint.Endpoint {
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		req := v.(*Request)
		resp := Response{
			Claims:       make(map[string]interface{}, len(req.Claims)),
			HeaderClaims: headerClaims,
		}
		token, err := f.NewToken(ctx, req, resp.Claims)
		if err != nil {
			return Response{}, err
		}

		resp.Body = []byte(token)

		return resp, err
	}
}

// NewClaimsEndpoint returns a go-kit endpoint that returns just the claims
func NewClaimsEndpoint(cb ClaimBuilder, headerClaims map[string]string) endpoint.Endpoint {
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		resp := Response{
			Claims:       make(map[string]interface{}),
			HeaderClaims: headerClaims,
		}
		if err := cb.AddClaims(ctx, v.(*Request), resp.Claims); err != nil {
			return nil, err
		}

		return resp.Claims, nil
	}
}

type Response struct {
	// Claims is the set of token claims.
	Claims map[string]interface{}
	// HeaderClaims is a map of claims-to-headers, where each claim will be attempted to be added as response header.
	HeaderClaims map[string]string
	// Body is the response body used by the EncodeIssueResponse.
	Body []byte
}

// Headers creates and returns a set of http headers based on HeaderClaims.
// Any failures to add claims are silent and does not affect the response.
func (resp Response) Headers() http.Header {
	headers := http.Header{}
	for claimKey, headerName := range resp.HeaderClaims {
		c, ok := resp.Claims[claimKey]
		if !ok {
			continue
		}

		switch v := c.(type) {
		case time.Time:
			headers.Add(headerName, v.Format(http.TimeFormat))
		case string:
			headers.Add(headerName, v)
		}
	}

	return headers
}
