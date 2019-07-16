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

func MessageKey() interface{} {
	return messageKey
}

func TimestampKey() interface{} {
	return timestampKey
}

func ErrorKey() interface{} {
	return errorKey
}

type Options struct {
	File       string
	Level      string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	JSON       bool
}

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

func New(o Options) (log.Logger, error) {
	if len(o.File) == 0 || o.File == StdoutFile {
		return Default(), nil
	}

	w := &lumberjack.Logger{
		Filename:   o.File,
		MaxSize:    o.MaxSize,
		MaxBackups: o.MaxBackups,
		MaxAge:     o.MaxAge,
	}

	var l log.Logger
	if o.JSON {
		l = log.NewJSONLogger(w)
	} else {
		l = log.NewLogfmtLogger(w)
	}

	l = log.WithPrefix(
		l,
		TimestampKey(), log.DefaultTimestampUTC,
	)

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

func Default() log.Logger {
	return defaultLogger
}
