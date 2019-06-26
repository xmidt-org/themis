package issuer

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

func NewEndpoint(i Issuer) endpoint.Endpoint {
	return func(_ context.Context, _ interface{}) (interface{}, error) {
		return i.Issue()
	}
}
