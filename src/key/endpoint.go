package key

import (
	"context"
	"errors"
	"xlog"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log/level"
)

var (
	ErrKeyNotFound = errors.New("That key does not exist")
)

func NewEndpoint(r Registry) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		pair, ok := r.Get(request.(string))
		xlog.Get(ctx).Log(
			level.Key(), level.InfoValue(),
			"pair", pair,
		)

		if !ok {
			return nil, ErrKeyNotFound
		}

		return pair, nil
	}
}
