package key

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xmidt-org/themis/xlog"
	"github.com/xmidt-org/themis/xlog/xlogtest"
)

func TestNewHandlerJWK(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)

			registry = NewRegistry(nil)
			endpoint = NewEndpoint(registry)
			handler  = NewHandlerJWK(endpoint)

			ctx     = xlog.With(context.Background(), xlogtest.New(t))
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

		data, err := ioutil.ReadAll(response.Body)
		require.NoError(err)
		require.NotEmpty(data)

		set, err := jwk.ParseBytes(data)
		require.NoError(err)
		assert.NotNil(set)
	})

	t.Run("NotFound", func(t *testing.T) {
		var (
			assert = assert.New(t)

			registry = NewRegistry(nil)
			endpoint = NewEndpoint(registry)
			handler  = NewHandlerJWK(endpoint)

			ctx     = xlog.With(context.Background(), xlogtest.New(t))
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

			ctx     = xlog.With(context.Background(), xlogtest.New(t))
			request = httptest.NewRequest("GET", "/", nil).WithContext(ctx)

			response = httptest.NewRecorder()
		)

		handler.ServeHTTP(response, request)
		assert.Equal(http.StatusInternalServerError, response.Code)
	})
}
