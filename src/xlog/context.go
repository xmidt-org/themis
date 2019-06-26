package xlog

import (
	"context"

	"github.com/go-kit/kit/log"
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
