package xlog

import (
	"github.com/go-kit/kit/log"
)

// keyUnmarshaller is the strategy for unmarshalling a logger.  It's declared here to avoid
// a circular dependency between packages.
type keyUnmarshaller interface {
	UnmarshalKey(string, interface{}) error
}

// Unmarshal loads an Options from a Viper instance and produces a go-kit Logger
func Unmarshal(key string, u keyUnmarshaller) (log.Logger, error) {
	var o Options
	if err := u.UnmarshalKey(key, &o); err != nil {
		return nil, err
	}

	return New(o)
}
