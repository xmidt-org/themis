package key

import (
	"context"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

type Handler http.Handler

func NewHandler(e endpoint.Endpoint) Handler {
	return kithttp.NewServer(
		e,
		func(_ context.Context, request *http.Request) (interface{}, error) {
			kid, ok := mux.Vars(request)["kid"]
			if !ok {
				return nil, ErrKeyNotFound
			}

			return kid, nil
		},
		func(_ context.Context, response http.ResponseWriter, value interface{}) error {
			response.Header().Set("Content-Type", "application/x-pem-file")
			_, err := value.(Pair).WriteVerifyPEMTo(response)
			return err
		},
		kithttp.ServerErrorEncoder(
			func(ctx context.Context, err error, response http.ResponseWriter) {
				if err == ErrKeyNotFound {
					response.WriteHeader(http.StatusNotFound)
					return
				}

				kithttp.DefaultErrorEncoder(ctx, err, response)
			},
		),
	)
}
