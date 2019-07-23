package xlog

import (
	"bytes"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvide(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		output   bytes.Buffer
		expected = log.NewJSONLogger(&output)
	)

	f := Provide(expected)
	require.NotNil(f)

	assert.Equal(expected, f())
}
