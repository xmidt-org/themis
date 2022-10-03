package xlog

import "github.com/go-kit/log"

// Provide takes a preexising logger and emits it as a component.  Useful when an external
// logger should be available in a container as a component.
func Provide(logger log.Logger) func() log.Logger {
	return func() log.Logger {
		return logger
	}
}
