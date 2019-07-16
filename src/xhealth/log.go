package xhealth

import (
	"fmt"
	"xlog"

	healthlog "github.com/InVisionApp/go-logger"
	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type loggerAdapter struct {
	kitlog.Logger
}

func (la loggerAdapter) Debug(msg ...interface{}) {
	la.Logger.Log(
		level.Key(), level.DebugValue(),
		xlog.MessageKey(), fmt.Sprint(msg...),
	)
}

func (la loggerAdapter) Info(msg ...interface{}) {
	la.Logger.Log(
		level.Key(), level.InfoValue(),
		xlog.MessageKey(), fmt.Sprint(msg...),
	)
}

func (la loggerAdapter) Warn(msg ...interface{}) {
	la.Logger.Log(
		level.Key(), level.WarnValue(),
		xlog.MessageKey(), fmt.Sprint(msg...),
	)
}

func (la loggerAdapter) Error(msg ...interface{}) {
	la.Logger.Log(
		level.Key(), level.ErrorValue(),
		xlog.MessageKey(), fmt.Sprint(msg...),
	)
}

func (la loggerAdapter) Debugln(msg ...interface{}) {
	la.Debug(msg...)
}

func (la loggerAdapter) Infoln(msg ...interface{}) {
	la.Info(msg...)
}

func (la loggerAdapter) Warnln(msg ...interface{}) {
	la.Warn(msg...)
}

func (la loggerAdapter) Errorln(msg ...interface{}) {
	la.Error(msg...)
}

func (la loggerAdapter) Debugf(format string, args ...interface{}) {
	la.Logger.Log(
		level.Key(), level.DebugValue(),
		xlog.MessageKey(), fmt.Sprintf(format, args...),
	)
}

func (la loggerAdapter) Infof(format string, args ...interface{}) {
	la.Logger.Log(
		level.Key(), level.InfoValue(),
		xlog.MessageKey(), fmt.Sprintf(format, args...),
	)
}

func (la loggerAdapter) Warnf(format string, args ...interface{}) {
	la.Logger.Log(
		level.Key(), level.WarnValue(),
		xlog.MessageKey(), fmt.Sprintf(format, args...),
	)
}

func (la loggerAdapter) Errorf(format string, args ...interface{}) {
	la.Logger.Log(
		level.Key(), level.ErrorValue(),
		xlog.MessageKey(), fmt.Sprintf(format, args...),
	)
}

func (la loggerAdapter) WithFields(f healthlog.Fields) healthlog.Logger {
	fields := make([]interface{}, 0, 2*len(f))
	for name, value := range f {
		fields = append(fields, name, value)
	}

	return loggerAdapter{
		Logger: kitlog.With(
			la.Logger,
			fields...,
		),
	}
}

// NewHealthLoggerAdapter adapts a go-kit logger onto the go-logger package's logger interface
func NewHealthLoggerAdapter(logger kitlog.Logger) healthlog.Logger {
	return loggerAdapter{Logger: logger}
}
