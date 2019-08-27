package xlog

import (
	"context"

	"github.com/go-kit/kit/log"
)

type contextKey struct{}

func Get(ctx context.Context) log.Logger {
	return GetDefault(ctx, Default())
}

func GetDefault(ctx context.Context, d log.Logger) log.Logger {
	l, ok := ctx.Value(contextKey{}).(log.Logger)
	if !ok {
		return d
	}

	return l
}

func With(ctx context.Context, l log.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, l)
}
