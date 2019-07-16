package xhealth

import (
	"context"

	health "github.com/InVisionApp/go-health"
	"github.com/go-kit/kit/log"
)

// Options holds the available configuration options for the health infrastructure
type Options struct {
	DisableLogging bool
	Custom         map[string]interface{}
}

// New constructs an IHealth instance for the given environment
func New(logger log.Logger, listener health.IStatusListener, o Options) (health.IHealth, error) {
	h := health.New()
	if o.DisableLogging || logger == nil {
		h.DisableLogging()
	} else {
		h.Logger = NewHealthLoggerAdapter(logger)
	}

	h.StatusListener = listener
	return h, nil
}

// OnStart returns an uber/fx Lifecycle hook for startup
func OnStart(h health.IHealth) func(context.Context) error {
	return func(_ context.Context) error {
		return h.Start()
	}
}

// OnStop returns an uber/fx Lifecycle hook for shutdown
func OnStop(h health.IHealth) func(context.Context) error {
	return func(_ context.Context) error {
		return h.Stop()
	}
}
