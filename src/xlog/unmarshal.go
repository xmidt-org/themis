package xlog

import (
	"xconfig"

	"github.com/go-kit/kit/log"
)

// Unmarshal loads an Options from a Viper instance and produces a go-kit Logger
func Unmarshal(key string, u xconfig.KeyUnmarshaller) (log.Logger, error) {
	var o Options
	if err := u.UnmarshalKey(key, &o); err != nil {
		return nil, err
	}

	return New(o)
}
