package xlog

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"go.uber.org/fx"
)

// Printer adapts a go-kit Logger onto the fx.Printer interface.  Use this type to
// direct uber/fx app logging to an externally created go-kit Logger.
type Printer struct {
	Logger log.Logger
}

func (p Printer) Printf(format string, parameters ...interface{}) {
	message := strings.ReplaceAll(
		fmt.Sprintf(format, parameters...),
		"\t", ": ",
	)

	p.Logger.Log(MessageKey(), message)
}

// DiscardPrinter is an fx.Printer that simply throws away all printed messages.
// This type is primarily useful for tests which don't need or want any output from uber/fx itself.
type DiscardPrinter struct{}

func (dp DiscardPrinter) Printf(string, ...interface{}) {
}

// BufferedPrinter is an uber/fx Printer that buffers log messages until a go-kit Logger
// is established.  This type is useful when a go-kit Logger is created as an uber/fx component
// and that Logger component should be used for all output.
type BufferedPrinter struct {
	lock     sync.Mutex
	messages []string
	logger   log.Logger
}

func (bp *BufferedPrinter) Printf(format string, parameters ...interface{}) {
	message := strings.ReplaceAll(
		fmt.Sprintf(format, parameters...),
		"\t", ": ",
	)

	defer bp.lock.Unlock()
	bp.lock.Lock()

	if bp.logger != nil {
		bp.logger.Log(MessageKey(), message)
	} else {
		bp.messages = append(bp.messages, message)
	}
}

// Len returns the count of messages currently buffered.  This will always be zero after
// SetLogger has been called.
func (bp *BufferedPrinter) Len() (c int) {
	bp.lock.Lock()
	c = len(bp.messages)
	bp.lock.Unlock()

	return
}

// SetLogger establishes the go-kit Logger that should receive uber/fx output.
// Any buffered messages are immediately to the supplied logger.  All calls to Printf
// afterward will not be buffered but sent to the supplied logger.
//
// This method is idempotent.  Subsequent attempts to set the logger have no effect
// and output will continue to the original logger.
func (bp *BufferedPrinter) SetLogger(l log.Logger) {
	defer bp.lock.Unlock()
	bp.lock.Lock()

	if bp.logger != nil {
		return
	}

	bp.logger = l
	for _, m := range bp.messages {
		bp.logger.Log(level.Key(), level.DebugValue(), MessageKey(), m)
	}

	bp.messages = nil
}

// OnStart ensures that any messages are flushed to avoid holding onto them during application execution
func (bp *BufferedPrinter) OnStart(context.Context) error {
	bp.lock.Lock()
	bp.messages = nil
	bp.lock.Unlock()

	return nil
}

// HandleError logs an error through the established logger.  If no logger has been
// set yet, the Error() logger is used.
func (bp *BufferedPrinter) HandleError(err error) {
	// ensure messages are flushed
	bp.SetLogger(Error())

	bp.logger.Log(
		level.Key(), level.ErrorValue(),
		ErrorKey(), err,
	)
}

// Logger is an analogue to the fx.Logger option.  This function creates a BufferedPrinter
// and emits it as a component, sets it as the uber/fx logger, and adds it as an error handler.
// Other code can express a dependency on a *BufferedPrinter and set the logger.
func Logger() fx.Option {
	bp := new(BufferedPrinter)
	return fx.Options(
		fx.Logger(bp),
		fx.Provide(
			func(lc fx.Lifecycle) *BufferedPrinter {
				lc.Append(fx.Hook{
					OnStart: bp.OnStart,
				})

				return bp
			},
		),
		fx.ErrorHook(bp),
	)
}
