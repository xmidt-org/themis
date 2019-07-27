package xlog

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// viperUnmarshaller is a locally defined KeyUnmarshaller to avoid a circular dependency
type viperUnmarshaller struct {
	v *viper.Viper
}

func (vu viperUnmarshaller) UnmarshalKey(key string, value interface{}) error {
	return vu.v.UnmarshalKey(key, value)
}

func TestUnmarshal(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		var (
			assert = assert.New(t)
			u      = viperUnmarshaller{v: viper.New()}
		)

		logger, err := Unmarshal("log", u)
		assert.NotNil(logger)
		assert.NoError(err)
	})

	t.Run("Invalid", func(t *testing.T) {
		var (
			assert = assert.New(t)
			v      = viper.New()
			u      = viperUnmarshaller{v: v}
		)

		v.Set("log.maxSize", "a;al;sdkfjal;dskfj")
		logger, err := Unmarshal("log", u)
		assert.Nil(logger)
		assert.Error(err)
	})
}
