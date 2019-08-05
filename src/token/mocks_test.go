package token

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type mockFactory struct {
	mock.Mock
}

func (m *mockFactory) NewToken(ctx context.Context, r *Request) (string, error) {
	arguments := m.Called(ctx, r)
	return arguments.String(0), arguments.Error(1)
}

func (m *mockFactory) ExpectNewToken(ctx context.Context, r *Request) *mock.Call {
	return m.On("NewToken", ctx, r)
}

type mockClaimBuilder struct {
	mock.Mock
}

func (m *mockClaimBuilder) AddClaims(ctx context.Context, r *Request, target map[string]interface{}) error {
	return m.Called(ctx, r, target).Error(0)
}

func (m *mockClaimBuilder) ExpectAddClaims(ctx context.Context, r *Request, target map[string]interface{}) *mock.Call {
	return m.On("AddClaims", ctx, r, target)
}
