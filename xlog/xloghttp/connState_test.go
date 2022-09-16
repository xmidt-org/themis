package xloghttp

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConnStateLogger(t *testing.T) {
	t.Run("NoLevel", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)

			output   bytes.Buffer
			original = log.NewJSONLogger(&output)

			connState = NewConnStateLogger(original, "connState", nil)
		)

		require.NotNil(connState)
		connState(nil, http.StateNew)
		assert.NotContains(output.String(), level.Key())
		assert.Contains(output.String(), "connState")
		assert.Contains(output.String(), http.StateNew.String())
	})

	t.Run("WithLevel", func(t *testing.T) {
		var (
			assert  = assert.New(t)
			require = require.New(t)

			output   bytes.Buffer
			original = log.NewJSONLogger(&output)

			connState = NewConnStateLogger(original, "connState", level.InfoValue())
		)

		require.NotNil(connState)
		connState(nil, http.StateNew)
		assert.Contains(output.String(), level.Key())
		assert.Contains(output.String(), level.InfoValue().String())
		assert.Contains(output.String(), "connState")
		assert.Contains(output.String(), http.StateNew.String())
	})
}
