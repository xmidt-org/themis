package xlogtest

import "github.com/go-kit/kit/log"

type test interface {
	Log(...interface{})
}

type loggerAdapter struct {
	t test
}

func (la loggerAdapter) Log(args ...interface{}) error {
	la.t.Log(args...)
	return nil
}

// New accepts either a *testing.T or a *testing.B and produces a go-kit logger that emits
// all it's name/value pairs as part of the test log.
func New(t test) log.Logger {
	return loggerAdapter{t: t}
}
