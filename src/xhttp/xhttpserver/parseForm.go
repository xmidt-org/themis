package xhttpserver

import (
	"net/http"

	kithttp "github.com/go-kit/kit/transport/http"
	"go.uber.org/fx"
)

// ParseForm handles invoking request.ParseForm and any error handling
type ParseForm struct {
	ErrorEncoder kithttp.ErrorEncoder
}

func (pf ParseForm) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if err := request.ParseForm(); err != nil {
			ee := pf.ErrorEncoder
			if ee == nil {
				ee = kithttp.DefaultErrorEncoder
			}

			ee(request.Context(), err, response)
			return
		}

		next.ServeHTTP(response, request)
	})
}

type ParseFormIn struct {
	fx.In

	ErrorEncoder kithttp.ErrorEncoder `optional:"true"`
}

func ProvideParseForm(in ParseFormIn) ParseForm {
	return ParseForm{ErrorEncoder: in.ErrorEncoder}
}
