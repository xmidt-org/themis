package xlog

import "github.com/go-kit/kit/log"

// Provide takes a preexising logger and emits it as a component.
// This allows bootstrapping of a logger prior to application creation.
func Provide(logger log.Logger) func() log.Logger {
	return func() log.Logger {
		return logger
	}
}
