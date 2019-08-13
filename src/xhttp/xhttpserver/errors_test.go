package xhttpserver

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMissingValueError(t *testing.T) {
	t.Run("Header", func(t *testing.T) {
		var (
			assert = assert.New(t)
			mve    = MissingValueError{
				Header: "X-Stuff",
			}
		)

		assert.Contains(mve.Error(), "X-Stuff")
		assert.Equal(http.StatusBadRequest, mve.StatusCode())
	})

	t.Run("HeaderOrParameter", func(t *testing.T) {
		var (
			assert = assert.New(t)
			mve    = MissingValueError{
				Header:    "X-Stuff",
				Parameter: "stuff",
			}
		)

		assert.Contains(mve.Error(), "X-Stuff")
		assert.Contains(mve.Error(), "stuff")
		assert.Equal(http.StatusBadRequest, mve.StatusCode())
	})

	t.Run("Parameter", func(t *testing.T) {
		var (
			assert = assert.New(t)
			mve    = MissingValueError{
				Parameter: "stuff",
			}
		)

		assert.Contains(mve.Error(), "stuff")
		assert.Equal(http.StatusBadRequest, mve.StatusCode())
	})
}

func TestMissingVariableError(t *testing.T) {
	var (
		assert = assert.New(t)
		mve    = MissingVariableError{
			Variable: "stuff",
		}
	)

	assert.Contains(mve.Error(), "stuff")
	assert.Equal(http.StatusInternalServerError, mve.StatusCode())
}
