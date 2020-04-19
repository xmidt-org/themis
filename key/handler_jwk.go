package key

import (
	"context"
	"net/http"

	"github.com/xmidt-org/themis/xlog"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log/level"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

type HandlerJWK http.Handler

func NewHandlerJWK(e endpoint.Endpoint) Handler {
	return kithttp.NewServer(
		e,
		func(ctx context.Context, request *http.Request) (interface{}, error) {
			kid, ok := mux.Vars(request)["kid"]
			if !ok {
				return nil, ErrNoKidVariable
			}

			xlog.Get(ctx).Log(
				level.Key(), level.InfoValue(),
				xlog.MessageKey(), "key request jwk",
				"kid", kid,
			)

			return kid, nil
		},
		func(_ context.Context, response http.ResponseWriter, value interface{}) error {
			response.Header().Set("Content-Type", "application/json")
			_, err := value.(Pair).WriteJWK(response)
			return err
		},
	)
}
