package token

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// NewServerEndpoint returns a go-kit endpoint for a single token Factory
func NewServerEndpoint(f Factory) endpoint.Endpoint {
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		return f.NewToken(ctx, v.(*Request))
	}
}
