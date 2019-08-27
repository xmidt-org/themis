package xlog

import (
	"bytes"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
)

func TestPrinter(t *testing.T) {
	var (
		assert = assert.New(t)

		output  bytes.Buffer
		logger  = log.NewJSONLogger(&output)
		printer = Printer{Logger: logger}
	)

	printer.Printf("test %d", 123)
	assert.Contains(output.String(), "test 123")
}
