// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package key

import (
	"context"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xmidt-org/sallust"

	"github.com/gorilla/mux"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)

			registry = NewRegistry(nil)
			endpoint = NewEndpoint(registry)
			handler  = NewHandler(endpoint)

			ctx     = sallust.With(context.Background(), sallust.Default())
			request = mux.SetURLVars(
				httptest.NewRequest("GET", "/", nil).WithContext(ctx),
				map[string]string{"kid": "test"},
			)

			response = httptest.NewRecorder()
		)

		_, err := registry.Register(Descriptor{Kid: "test"})
		require.NoError(err)

		handler.ServeHTTP(response, request)
		assert.Equal(http.StatusOK, response.Code)
		assert.Equal("application/x-pem-file", response.Header().Get("Content-Type"))

		data, err := io.ReadAll(response.Body)
		require.NoError(err)
		require.NotEmpty(data)

		block, _ := pem.Decode(data)
		assert.NotNil(block)
	})

	t.Run("NotFound", func(t *testing.T) {
		var (
			assert = assert.New(t)

			registry = NewRegistry(nil)
			endpoint = NewEndpoint(registry)
			handler  = NewHandler(endpoint)

			ctx     = sallust.With(context.Background(), sallust.Default())
			request = mux.SetURLVars(
				httptest.NewRequest("GET", "/", nil).WithContext(ctx),
				map[string]string{"kid": "test"},
			)

			response = httptest.NewRecorder()
		)

		handler.ServeHTTP(response, request)
		assert.Equal(http.StatusNotFound, response.Code)
	})

	t.Run("NoKidVariable", func(t *testing.T) {
		var (
			assert = assert.New(t)

			registry = NewRegistry(nil)
			endpoint = NewEndpoint(registry)
			handler  = NewHandler(endpoint)

			ctx     = sallust.With(context.Background(), sallust.Default())
			request = httptest.NewRequest("GET", "/", nil).WithContext(ctx)

			response = httptest.NewRecorder()
		)

		handler.ServeHTTP(response, request)
		assert.Equal(http.StatusInternalServerError, response.Code)
	})
}

func TestNewHandlerJWK(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)

			registry = NewRegistry(nil)
			endpoint = NewEndpoint(registry)
			handler  = NewHandlerJWK(endpoint)

			ctx     = sallust.With(context.Background(), sallust.Default())
			request = mux.SetURLVars(
				httptest.NewRequest("GET", "/", nil).WithContext(ctx),
				map[string]string{"kid": "test"},
			)

			response = httptest.NewRecorder()
		)

		_, err := registry.Register(Descriptor{Kid: "test"})
		require.NoError(err)

		handler.ServeHTTP(response, request)
		assert.Equal(http.StatusOK, response.Code)
		assert.Equal("application/json", response.Header().Get("Content-Type"))

		data, err := io.ReadAll(response.Body)
		require.NoError(err)
		require.NotEmpty(data)

		set, err := jwk.Parse(data)
		require.NoError(err)
		assert.NotNil(set)
	})

	t.Run("NotFound", func(t *testing.T) {
		var (
			assert = assert.New(t)

			registry = NewRegistry(nil)
			endpoint = NewEndpoint(registry)
			handler  = NewHandlerJWK(endpoint)

			ctx     = sallust.With(context.Background(), sallust.Default())
			request = mux.SetURLVars(
				httptest.NewRequest("GET", "/", nil).WithContext(ctx),
				map[string]string{"kid": "test"},
			)

			response = httptest.NewRecorder()
		)

		handler.ServeHTTP(response, request)
		assert.Equal(http.StatusNotFound, response.Code)
	})

	t.Run("NoKidVariable", func(t *testing.T) {
		var (
			assert = assert.New(t)

			registry = NewRegistry(nil)
			endpoint = NewEndpoint(registry)
			handler  = NewHandlerJWK(endpoint)

			ctx     = sallust.With(context.Background(), sallust.Default())
			request = httptest.NewRequest("GET", "/", nil).WithContext(ctx)

			response = httptest.NewRecorder()
		)

		handler.ServeHTTP(response, request)
		assert.Equal(http.StatusInternalServerError, response.Code)
	})
}
