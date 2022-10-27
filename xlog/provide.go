package xlog

import (
	"go.uber.org/zap"
)

// Provide takes a preexising logger and emits it as a component.  Useful when an external
// logger should be available in a container as a component.
func Provide(logger *zap.Logger) func() *zap.Logger {
	return func() *zap.Logger {
		return logger
	}
}
