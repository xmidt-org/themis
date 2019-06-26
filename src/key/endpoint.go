package key

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"
)

var (
	ErrKeyNotFound = errors.New("That key does not exist")
)

func NewEndpoint(r Registry) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		pair, ok := r.Get(request.(string))
		if !ok {
			return nil, ErrKeyNotFound
		}

		return pair, nil
	}
}
