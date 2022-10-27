package xlog

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testAllowLevelDebug(t *testing.T, value string) {
	var (
		assert = assert.New(t)

		output   bytes.Buffer
		original = log.NewJSONLogger(&output)
	)

	logger, err := AllowLevel(original, value)
	assert.Equal(original, logger)
	assert.NoError(err)
}

func testAllowLevelInfo(t *testing.T, value string) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		output   bytes.Buffer
		original = log.NewJSONLogger(&output)
	)

	logger, err := AllowLevel(original, value)
	require.NotEqual(original, logger)
	require.NoError(err)

	output.Reset()
	logger.Log(level.Key(), level.DebugValue(), "test", "test")
	assert.Zero(output.Len())

	output.Reset()
	logger.Log(level.Key(), level.InfoValue(), "test", "test")
	assert.True(output.Len() > 0)

	output.Reset()
	logger.Log(level.Key(), level.WarnValue(), "test", "test")
	assert.True(output.Len() > 0)

	output.Reset()
	logger.Error("test", "test")
	assert.True(output.Len() > 0)
}

func testAllowLevelWarn(t *testing.T, value string) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		output   bytes.Buffer
		original = log.NewJSONLogger(&output)
	)

	logger, err := AllowLevel(original, value)
	require.NotEqual(original, logger)
	require.NoError(err)

	output.Reset()
	logger.Log(level.Key(), level.DebugValue(), "test", "test")
	assert.Zero(output.Len())

	output.Reset()
	logger.Log(level.Key(), level.InfoValue(), "test", "test")
	assert.Zero(output.Len())

	output.Reset()
	logger.Log(level.Key(), level.WarnValue(), "test", "test")
	assert.True(output.Len() > 0)

	output.Reset()
	logger.Error("test", "test")
	assert.True(output.Len() > 0)
}

func testAllowLevelError(t *testing.T, value string) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		output   bytes.Buffer
		original = log.NewJSONLogger(&output)
	)

	logger, err := AllowLevel(original, value)
	require.NotEqual(original, logger)
	require.NoError(err)

	output.Reset()
	logger.Log(level.Key(), level.DebugValue(), "test", "test")
	assert.Zero(output.Len())

	output.Reset()
	logger.Log(level.Key(), level.InfoValue(), "test", "test")
	assert.Zero(output.Len())

	output.Reset()
	logger.Log(level.Key(), level.WarnValue(), "test", "test")
	assert.Zero(output.Len())

	output.Reset()
	logger.Error("test", "test")
	assert.True(output.Len() > 0)
}

func testAllowLevelInvalid(t *testing.T, value string) {
	var (
		assert = assert.New(t)

		output   bytes.Buffer
		original = log.NewJSONLogger(&output)
	)

	logger, err := AllowLevel(original, value)
	assert.Equal(original, logger)
	assert.Error(err)
}

func TestAllowLevel(t *testing.T) {
	t.Run(LevelDebug, func(t *testing.T) {
		for _, v := range []string{"", LevelDebug, "debug"} {
			t.Run(v, func(t *testing.T) { testAllowLevelDebug(t, v) })
		}
	})

	t.Run(LevelInfo, func(t *testing.T) {
		for _, v := range []string{LevelInfo, "info"} {
			t.Run(v, func(t *testing.T) { testAllowLevelInfo(t, v) })
		}
	})

	t.Run(LevelWarn, func(t *testing.T) {
		for _, v := range []string{LevelWarn, "warn"} {
			t.Run(v, func(t *testing.T) { testAllowLevelWarn(t, v) })
		}
	})

	t.Run(LevelError, func(t *testing.T) {
		for _, v := range []string{LevelError, "error"} {
			t.Run(v, func(t *testing.T) { testAllowLevelError(t, v) })
		}
	})

	t.Run("invalid", func(t *testing.T) {
		for _, v := range []string{"unrecognized", "($*&)@(&#"} {
			t.Run(v, func(t *testing.T) { testAllowLevelInvalid(t, v) })
		}
	})
}

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

	for _, record := range testData {
		t.Run(record.value, func(t *testing.T) {
			assert := assert.New(t)
			v, err := Level(record.value)
			assert.Equal(record.expected, v)
			assert.Equal(record.err, err != nil)
		})
	}
}

func testNewDefault(t *testing.T, o Options) {
	assert := assert.New(t)

	logger, err := New(o)
	assert.Equal(defaultLogger, logger)
	assert.NoError(err)
}

func testNewValid(t *testing.T, o Options) {
	assert := assert.New(t)

	logger, err := New(o)
	assert.NotNil(logger)
	assert.NoError(err)
}

func testNewInvalid(t *testing.T, o Options) {
	assert := assert.New(t)

	logger, err := New(o)
	assert.Nil(logger)
	assert.Error(err)
}

func TestNew(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		testData := []Options{
			Options{},
			Options{File: StdoutFile},
		}

		for i, o := range testData {
			t.Run(strconv.Itoa(i), func(t *testing.T) { testNewDefault(t, o) })
		}
	})

	t.Run("Valid", func(t *testing.T) {
		testData := []Options{
			Options{File: "test.log", JSON: false},
			Options{File: "test.log", Level: "INFO", JSON: false},
			Options{File: "test.log", JSON: true},
			Options{File: "test.log", Level: "INFO", JSON: true},
		}

		for i, o := range testData {
			t.Run(strconv.Itoa(i), func(t *testing.T) { testNewValid(t, o) })
		}
	})

	t.Run("Invalid", func(t *testing.T) {
		testData := []Options{
			Options{Level: "invalid"},
			Options{File: "test.log", Level: "invalid"},
		}

		for i, o := range testData {
			t.Run(strconv.Itoa(i), func(t *testing.T) { testNewInvalid(t, o) })
		}
	})
}

func TestDefault(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(defaultLogger, Default())
}
