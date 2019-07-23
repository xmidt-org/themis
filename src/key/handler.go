package key

import (
	"context"
	"errors"
	"net/http"
	"xlog"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log/level"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

var (
	ErrNoKidVariable = errors.New("No kid variable in URI definition")
)

type Handler http.Handler

func NewHandler(e endpoint.Endpoint) Handler {
	return kithttp.NewServer(
		e,
		func(ctx context.Context, request *http.Request) (interface{}, error) {
			kid, ok := mux.Vars(request)["kid"]
			if !ok {
				return nil, ErrNoKidVariable
			}

			xlog.Get(ctx).Log(
				level.Key(), level.InfoValue(),
				xlog.MessageKey(), "key request",
				"kid", kid,
			)

			return kid, nil
		},
		func(_ context.Context, response http.ResponseWriter, value interface{}) error {
			response.Header().Set("Content-Type", "application/x-pem-file")
			_, err := value.(Pair).WriteVerifyPEMTo(response)
			return err
		},
	)
}
