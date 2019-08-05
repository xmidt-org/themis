package token

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func testNewIssueEndpointSuccess(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		factory  = new(mockFactory)
		request  = NewRequest()
		endpoint = NewIssueEndpoint(factory)
	)

	require.NotNil(endpoint)
	factory.ExpectNewToken(context.Background(), request).Once().Return("test", error(nil))
	token, err := endpoint(context.Background(), request)
	assert.Equal("test", token)
	assert.NoError(err)

	factory.AssertExpectations(t)
}

func testNewIssueEndpointFailure(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		factory     = new(mockFactory)
		expectedErr = errors.New("expected")
		request     = NewRequest()
		endpoint    = NewIssueEndpoint(factory)
	)

	require.NotNil(endpoint)
	factory.ExpectNewToken(context.Background(), request).Once().Return("", expectedErr)
	token, actualErr := endpoint(context.Background(), request)
	assert.Equal("", token)
	assert.Equal(expectedErr, actualErr)

	factory.AssertExpectations(t)
}

func TestNewIssueEndpoint(t *testing.T) {
	t.Run("Success", testNewIssueEndpointSuccess)
	t.Run("Failure", testNewIssueEndpointFailure)
}

func testNewClaimsEndpointSuccess(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		builder        = new(mockClaimBuilder)
		expectedClaims = map[string]interface{}{"key": "value"}
		request        = NewRequest()
		endpoint       = NewClaimsEndpoint(builder)
	)

	require.NotNil(endpoint)
	builder.ExpectAddClaims(context.Background(), request, map[string]interface{}{}).Once().Return(error(nil)).
		Run(func(arguments mock.Arguments) {
			arguments.Get(2).(map[string]interface{})["key"] = "value"
		})

	actualClaims, err := endpoint(context.Background(), request)
	assert.Equal(expectedClaims, actualClaims)
	assert.NoError(err)

	builder.AssertExpectations(t)
}

func testNewClaimsEndpointFailure(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		builder     = new(mockClaimBuilder)
		expectedErr = errors.New("expected")
		request     = NewRequest()
		endpoint    = NewClaimsEndpoint(builder)
	)

	require.NotNil(endpoint)
	builder.ExpectAddClaims(context.Background(), request, map[string]interface{}{}).Once().Return(expectedErr)
	claims, actualErr := endpoint(context.Background(), request)
	assert.Empty(claims)
	assert.Equal(expectedErr, actualErr)

	builder.AssertExpectations(t)
}

func TestNewClaimsEndpoint(t *testing.T) {
	t.Run("Success", testNewClaimsEndpointSuccess)
	t.Run("Failure", testNewClaimsEndpointFailure)
}
