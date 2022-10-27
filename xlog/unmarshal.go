package xlog

import (
	"github.com/xmidt-org/themis/config"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// LogUnmarshalIn defines the set of dependencies for unmarshalling a go-kit logger
type LogUnmarshalIn struct {
	fx.In

	// Unmarshaller is the required strategy for unmarshalling an Options
	Unmarshaller config.Unmarshaller

	// Printer is the optional BufferedPrinter component.  If present, the unmarshalled logger
	// will be set as this printer's logger.
	Printer *BufferedPrinter `optional:"true"`
}

// Unmarshal returns an uber/fx provider function that handles unmarshalling a logger and emitted it as a component.
// If a *BufferedPrinter component is present, the unmarshalled logger will be set as that printer's logger.
func Unmarshal(key string) func(LogUnmarshalIn) (*zap.Logger, error) {
	return func(in LogUnmarshalIn) (*zap.Logger, error) {
		var o Options
		if err := in.Unmarshaller.UnmarshalKey(key, &o); err != nil {
			return nil, err
		}

		l, err := New(o)
		if err == nil && in.Printer != nil {
			in.Printer.SetLogger(l)
		}

		return l, err
	}
}
