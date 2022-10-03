package xlog

import (
	"bytes"
	"context"
	"testing"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	t.Run("InContext", func(t *testing.T) {
		var (
			assert = assert.New(t)

			output   bytes.Buffer
			expected = log.NewJSONLogger(&output)
			ctx      = context.WithValue(context.Background(), contextKey{}, expected)
		)

		assert.Equal(expected, Get(ctx))
	})

	t.Run("NotInContext", func(t *testing.T) {
		assert := assert.New(t)
		assert.Equal(defaultLogger, Get(context.Background()))
	})
}

func TestGetDefault(t *testing.T) {
	t.Run("NilDefault", func(t *testing.T) {
		t.Run("InContext", func(t *testing.T) {
			var (
				assert = assert.New(t)

				output   bytes.Buffer
				expected = log.NewJSONLogger(&output)
				ctx      = context.WithValue(context.Background(), contextKey{}, expected)
			)

			assert.Equal(expected, GetDefault(ctx, nil))
		})

		t.Run("NotInContext", func(t *testing.T) {
			assert := assert.New(t)
			assert.Nil(GetDefault(context.Background(), nil))
		})
	})

	t.Run("WithDefault", func(t *testing.T) {
		t.Run("InContext", func(t *testing.T) {
			var (
				assert = assert.New(t)

				output   bytes.Buffer
				d        = log.NewLogfmtLogger(&output)
				expected = log.NewJSONLogger(&output)
				ctx      = context.WithValue(context.Background(), contextKey{}, expected)
			)

			assert.Equal(expected, GetDefault(ctx, d))
		})

		t.Run("NotInContext", func(t *testing.T) {
			var (
				assert = assert.New(t)

				output bytes.Buffer
				d      = log.NewLogfmtLogger(&output)
			)

			assert.Equal(d, GetDefault(context.Background(), d))
		})
	})
}

func TestWith(t *testing.T) {
	var (
		assert = assert.New(t)

		output   bytes.Buffer
		expected = log.NewJSONLogger(&output)
		ctx      = With(context.Background(), expected)
	)

	assert.Equal(expected, ctx.Value(contextKey{}))
}
