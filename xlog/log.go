package xlog

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

const (
	StdoutFile = "stdout"

	messageKey   = "msg"
	timestampKey = "ts"
	errorKey     = "error"

	LevelNone  = ""
	LevelError = "ERROR"
	LevelWarn  = "WARN"
	LevelInfo  = "INFO"
	LevelDebug = "DEBUG"
)

// MessageKey returns the logging key for an arbitrary message
func MessageKey() interface{} {
	return messageKey
}

// TimestampKey returns the logging key for the timestamp
func TimestampKey() interface{} {
	return timestampKey
}

// ErrorKey returns the logging key for error text
func ErrorKey() interface{} {
	return errorKey
}

// Options defines the set of configuration options for a go-kit log.LoggeAr.
type Options struct {
	// File is the output destination for logs.  If unset or set to StdoutFile,
	// a console logger is created and the log rolling options are ignored.
	File string

	// Level is the max logging level for output.
	Level string

	// MaxSize is the lumberjack maximum size when rolling logs
	MaxSize int

	// MaxBackups is the lumberjack maximum backs when rolling logs
	MaxBackups int

	// MaxAge is the lumberjack maximum age when rolling logs
	MaxAge int

	// JSON indicates whether a go-kit JSON logger or a logfmt logger is used
	JSON bool
}

// AllowLevel produces a filtered logger with the given level.AllowXXX set.
func AllowLevel(next log.Logger, v string) (log.Logger, error) {
	switch strings.ToUpper(v) {
	case LevelNone:
		fallthrough
	case LevelDebug:
		// optimization: debug allows everything, so no reason to waste resources
		// with a filtered logger
		return next, nil

	case LevelError:
		return level.NewFilter(next, level.AllowError()), nil

	case LevelWarn:
		return level.NewFilter(next, level.AllowWarn()), nil

	case LevelInfo:
		return level.NewFilter(next, level.AllowInfo()), nil

	default:
		return next, fmt.Errorf("Unrecognized log level: %s", v)
	}
}

// Level returns the go-kit level.Value for a given configuration string value.
func Level(v string) (level.Value, error) {
	switch strings.ToUpper(v) {
	case LevelNone:
		return nil, nil

	case LevelDebug:
		return level.DebugValue(), nil

	case LevelError:
		return level.ErrorValue(), nil

	case LevelWarn:
		return level.WarnValue(), nil

	case LevelInfo:
		return level.InfoValue(), nil

	default:
		return nil, fmt.Errorf("Unrecognized log level: %s", v)
	}
}

// New produces a go-kit log.Logger using the given set of configuration options
func New(o Options) (log.Logger, error) {
	var l log.Logger
	if len(o.File) == 0 || o.File == StdoutFile {
		l = Default()
	} else {
		w := &lumberjack.Logger{
			Filename:   o.File,
			MaxSize:    o.MaxSize,
			MaxBackups: o.MaxBackups,
			MaxAge:     o.MaxAge,
		}

		if o.JSON {
			l = log.NewJSONLogger(w)
		} else {
			l = log.NewLogfmtLogger(w)
		}

		l = log.WithPrefix(
			l,
			TimestampKey(), log.DefaultTimestampUTC,
		)
	}

	if levelled, err := AllowLevel(l, o.Level); err != nil {
		return nil, err
	} else {
		l = levelled
	}

	return l, nil
}

var defaultLogger = log.WithPrefix(
	log.NewJSONLogger(
		log.NewSyncWriter(os.Stdout),
	),
	TimestampKey(), log.DefaultTimestampUTC,
)

// Default() returns the singleton default go-kit logger, which writes to stdout and
// is safe for concurrent usage.
func Default() log.Logger {
	return defaultLogger
}

var errorLogger = log.WithPrefix(
	log.NewJSONLogger(
		log.NewSyncWriter(os.Stderr),
	),
	TimestampKey(), log.DefaultTimestampUTC,
)

// Error returns a default go-kit Logger that sends all output to os.Stderr.  Useful
// for reporting errors on a crash.
func Error() log.Logger {
	return errorLogger
}

type discardLogger struct{}

func (dl discardLogger) Log(...interface{}) error {
	return nil
}

// Discard() returns a go-kit logger that simply ignores any calls to Log
func Discard() log.Logger {
	return discardLogger{}
}
