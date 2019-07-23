package xlog

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshal(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		var (
			assert = assert.New(t)

			v = viper.New()
		)

		logger, err := Unmarshal("log", v)
		assert.NotNil(logger)
		assert.NoError(err)
	})

	t.Run("Invalid", func(t *testing.T) {
		var (
			assert = assert.New(t)

			v = viper.New()
		)

		v.Set("log.maxSize", "a;al;sdkfjal;dskfj")
		logger, err := Unmarshal("log", v)
		assert.Nil(logger)
		assert.Error(err)
	})
}
