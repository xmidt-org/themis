package issuer

import (
	"context"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
)

type Handler http.Handler

func NewHandler(e endpoint.Endpoint) Handler {
	return kithttp.NewServer(
		e,
		func(_ context.Context, _ *http.Request) (interface{}, error) {
			return struct{}{}, nil
		},
		func(_ context.Context, response http.ResponseWriter, value interface{}) error {
			response.Header().Set("Content-Type", "application/jose")
			_, err := response.Write([]byte(value.(string)))
			return err
		},
	)
}
