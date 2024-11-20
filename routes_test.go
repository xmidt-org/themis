// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/xmidt-org/themis/key"
)

func TestBuildKeyRoutes(t *testing.T) {
	var (
		handlerPEM = http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			kid := mux.Vars(request)["kid"]
			response.Header().Set("Content-Type", key.ContentTypePEM)
			if len(kid) == 0 {
				response.WriteHeader(http.StatusInternalServerError)
			}

			response.Write([]byte("pem"))
		})

		handlerJWK = http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			kid := mux.Vars(request)["kid"]
			response.Header().Set("Content-Type", key.ContentTypeJWK)
			if len(kid) == 0 {
				response.WriteHeader(http.StatusInternalServerError)
			}

			response.Write([]byte("jwk"))
		})

		router = mux.NewRouter()
	)

	BuildKeyRoutes(KeyRoutesIn{
		Router:     router,
		Handler:    handlerPEM,
		HandlerJWK: handlerJWK,
	})

	t.Run("Default", func(t *testing.T) {
		var (
			assert   = assert.New(t)
			response = httptest.NewRecorder()
			request  = httptest.NewRequest("GET", "/keys/test", nil)
		)

		router.ServeHTTP(response, request)
		assert.Equal(http.StatusOK, response.Code)
		assert.Equal(key.ContentTypePEM, response.Header().Get("Content-Type"))
		assert.Equal("pem", response.Body.String())
	})

	t.Run("key.pem", func(t *testing.T) {
		var (
			assert   = assert.New(t)
			response = httptest.NewRecorder()
			request  = httptest.NewRequest("GET", "/keys/test/key.pem", nil)
		)

		router.ServeHTTP(response, request)
		assert.Equal(http.StatusOK, response.Code)
		assert.Equal(key.ContentTypePEM, response.Header().Get("Content-Type"))
		assert.Equal("pem", response.Body.String())
	})

	t.Run("key.json", func(t *testing.T) {
		var (
			assert   = assert.New(t)
			response = httptest.NewRecorder()
			request  = httptest.NewRequest("GET", "/keys/test/key.json", nil)
		)

		router.ServeHTTP(response, request)
		assert.Equal(http.StatusOK, response.Code)
		assert.Equal(key.ContentTypeJWK, response.Header().Get("Content-Type"))
		assert.Equal("jwk", response.Body.String())
	})

	t.Run("Accept", func(t *testing.T) {
		t.Run(key.ContentTypePEM, func(t *testing.T) {
			var (
				assert   = assert.New(t)
				response = httptest.NewRecorder()
				request  = httptest.NewRequest("GET", "/keys/test", nil)
			)

			request.Header.Set("Accept", key.ContentTypePEM)
			router.ServeHTTP(response, request)
			assert.Equal(http.StatusOK, response.Code)
			assert.Equal(key.ContentTypePEM, response.Header().Get("Content-Type"))
			assert.Equal("pem", response.Body.String())
		})

		t.Run(key.ContentTypeJWK, func(t *testing.T) {
			var (
				assert   = assert.New(t)
				response = httptest.NewRecorder()
				request  = httptest.NewRequest("GET", "/keys/test", nil)
			)

			request.Header.Set("Accept", key.ContentTypeJWK)
			router.ServeHTTP(response, request)
			assert.Equal(http.StatusOK, response.Code)
			assert.Equal(key.ContentTypeJWK, response.Header().Get("Content-Type"))
			assert.Equal("jwk", response.Body.String())
		})
	})
}
