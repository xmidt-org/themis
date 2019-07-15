package token

import (
	"net/http"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
)

type Handler http.Handler

func NewHandler(e endpoint.Endpoint, b ...RequestBuilder) Handler {
	return kithttp.NewServer(
		e,
		DecodeServerRequest(b...),
		EncodeServerResponse,
	)
}
