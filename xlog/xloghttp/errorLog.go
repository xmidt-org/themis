package xloghttp

import (
	stdlibLog "log"

	"github.com/go-kit/kit/log"
)

func NewErrorLog(address string, logger log.Logger) *stdlibLog.Logger {
	return stdlibLog.New(
		log.NewStdlibAdapter(logger),
		address,
		stdlibLog.LstdFlags|stdlibLog.LUTC,
	)
}
