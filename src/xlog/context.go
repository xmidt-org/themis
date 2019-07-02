package xlog

import (
	"context"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/justinas/alice"
)

type contextKey struct{}

func Get(ctx context.Context) log.Logger {
	l, ok := ctx.Value(contextKey{}).(log.Logger)
	if !ok {
		return Default()
	}

	return l
}

func With(ctx context.Context, l log.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, l)
}

func WithRequest(original *http.Request, l log.Logger) *http.Request {
	ctx := With(
		original.Context(),
		log.WithPrefix(
			l,
			"requestMethod", original.Method,
			"requestURI", original.RequestURI,
			"remoteAddr", original.RemoteAddr,
		),
	)

	return original.WithContext(ctx)
}

func NewConstructor(l log.Logger) alice.Constructor {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			next.ServeHTTP(
				response,
				WithRequest(request, l),
			)
		})
	}
}
