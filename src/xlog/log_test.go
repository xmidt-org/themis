package xlog

import (
	"strconv"
	"testing"

	"github.com/go-kit/kit/log/level"
	"github.com/stretchr/testify/assert"
)

func TestLevel(t *testing.T) {
	testData := []struct {
		value    string
		expected level.Value
		err      bool
	}{
		{value: ""},
		{value: LevelDebug, expected: level.DebugValue()},
		{value: "debug", expected: level.DebugValue()},
		{value: LevelInfo, expected: level.InfoValue()},
		{value: "info", expected: level.InfoValue()},
		{value: LevelWarn, expected: level.WarnValue()},
		{value: "warn", expected: level.WarnValue()},
		{value: LevelError, expected: level.ErrorValue()},
		{value: "error", expected: level.ErrorValue()},
		{value: "unrecognized", err: true},
	}

	for i, record := range testData {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert := assert.New(t)
			v, err := Level(record.value)
			assert.Equal(record.expected, v)
			assert.Equal(record.err, err != nil)
		})
	}
}

func TestNew(t *testing.T) {
}
