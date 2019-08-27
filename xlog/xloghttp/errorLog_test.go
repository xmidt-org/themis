package xloghttp

import (
	"bytes"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewErrorLog(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		output bytes.Buffer
		logger = log.NewJSONLogger(&output)

		errorLog = NewErrorLog("foobar.com", logger)
	)

	require.NotNil(errorLog)
	errorLog.Print("hello")
	assert.Contains(output.String(), "foobar.com")
}
