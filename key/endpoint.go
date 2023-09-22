// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package key

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-kit/kit/endpoint"
)

type KeyNotFoundError struct {
	Kid string
}

func (knfe KeyNotFoundError) Error() string {
	return fmt.Sprintf("No key exists with kid %s", knfe.Kid)
}

func (knfe KeyNotFoundError) StatusCode() int {
	return http.StatusNotFound
}

func NewEndpoint(r Registry) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		kid := request.(string)
		pair, ok := r.Get(request.(string))
		if !ok {
			return nil, KeyNotFoundError{Kid: kid}
		}

		return pair, nil
	}
}
