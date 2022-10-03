package key

import (
	"context"
	"errors"
	"net/http"

	"github.com/xmidt-org/themis/xlog"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
)

const (
	ContentTypePEM = "application/x-pem-file"
	ContentTypeJWK = "application/json"
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
			response.Header().Set("Content-Type", ContentTypePEM)
			_, err := value.(Pair).WriteVerifyPEMTo(response)
			return err
		},
	)
}

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
			response.Header().Set("Content-Type", ContentTypeJWK)
			_, err := value.(Pair).WriteJWK(response)
			return err
		},
	)
}
