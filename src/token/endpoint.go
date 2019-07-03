package token

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

func NewEndpoint(f Factory) endpoint.Endpoint {
	return func(_ context.Context, v interface{}) (interface{}, error) {
		return f.NewToken(v.(*Request))
	}
}
