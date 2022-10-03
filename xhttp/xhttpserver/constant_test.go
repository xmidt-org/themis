package xhttpserver

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConstantHandler(t *testing.T) {
	testData := []struct {
		constant           Constant
		expectedStatusCode int
		expectedHeader     http.Header
		expectedBody       []byte
	}{
		{
			constant:           Constant{},
			expectedStatusCode: http.StatusOK,
			expectedHeader: http.Header{
				"Content-Length": []string{"0"},
			},
		},
		{
			constant: Constant{
				StatusCode: 516,
				Header: http.Header{
					"X-Custom1": []string{"value1"},
					"x-CUStom2": []string{"value1", "value2"},
				},
				Body: []byte("hello, world"),
			},
			expectedStatusCode: 516,
			expectedHeader: http.Header{
				"Content-Length": []string{strconv.Itoa(len("hello, world"))},
				"X-Custom1":      []string{"value1"},
				"X-Custom2":      []string{"value1", "value2"},
			},
			expectedBody: []byte("hello, world"),
		},
	}

	for i, record := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var (
				assert     = assert.New(t)
				require    = require.New(t)
				actualBody bytes.Buffer

				handler  = record.constant.NewHandler()
				response = httptest.NewRecorder()
				request  = httptest.NewRequest("GET", "/", nil)
			)

			require.NotNil(handler)
			response.Body = &actualBody
			handler.ServeHTTP(response, request)
			assert.Equal(response.Code, record.expectedStatusCode)
			assert.Equal(response.Header(), record.expectedHeader)
			assert.Equal(record.expectedBody, actualBody.Bytes())
		})
	}
}
