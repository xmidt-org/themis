package token

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// NewIssueEndpoint returns a go-kit endpoint for a token factory's NewToken method
func NewIssueEndpoint(f Factory) endpoint.Endpoint {
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		return f.NewToken(ctx, v.(*Request))
	}
}

// NewClaimsEndpoint returns a go-kit endpoint that returns just the claims
func NewClaimsEndpoint(cb ClaimBuilder) endpoint.Endpoint {
	return func(ctx context.Context, v interface{}) (interface{}, error) {
		merged := make(map[string]interface{})
		if err := cb.Append(ctx, v.(*Request), merged); err != nil {
			return nil, err
		}

		return merged, nil
	}
}
