package xlog

import (
	"fmt"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// Printer is an uber/fx printer for the application.  It adapts a go-kit logger onto the fx.Printer interface.
type Printer struct {
	log.Logger
}

func (p Printer) Printf(format string, parameters ...interface{}) {
	message := fmt.Sprintf(format, parameters...)
	p.Logger.Log(
		level.Key(), level.InfoValue(),
		MessageKey(), strings.ReplaceAll(message, "\t", ": "),
	)
}
