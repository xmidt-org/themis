package xlog

import (
	"testing"
	"xconfig"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshal(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		var (
			assert = assert.New(t)
			u      = xconfig.ViperUnmarshaller{Viper: viper.New()}
		)

		logger, err := Unmarshal("log", u)
		assert.NotNil(logger)
		assert.NoError(err)
	})

	t.Run("Invalid", func(t *testing.T) {
		var (
			assert = assert.New(t)
			v      = viper.New()
			u      = xconfig.ViperUnmarshaller{Viper: v}
		)

		v.Set("log.maxSize", "a;al;sdkfjal;dskfj")
		logger, err := Unmarshal("log", u)
		assert.Nil(logger)
		assert.Error(err)
	})
}
