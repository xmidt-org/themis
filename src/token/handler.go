package token

import (
	"net/http"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
)

type IssueHandler http.Handler

func NewIssueHandler(e endpoint.Endpoint, rb RequestBuilders) IssueHandler {
	return kithttp.NewServer(
		e,
		DecodeServerRequest(rb),
		EncodeServerResponse,
	)
}

type ClaimsHandler http.Handler

func NewClaimsHandler(e endpoint.Endpoint, rb RequestBuilders) ClaimsHandler {
	return kithttp.NewServer(
		e,
		DecodeServerRequest(rb),
		EncodeServerResponse,
	)
}
