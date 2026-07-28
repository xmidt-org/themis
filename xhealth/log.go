// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package xhealth

import (
	"fmt"

	healthlog "github.com/InVisionApp/go-logger"
	"go.uber.org/zap"
)

type loggerAdapter struct {
	Logger *zap.Logger
}

func (la loggerAdapter) Debug(msg ...any) {
	la.Logger.Debug(fmt.Sprint(msg...))
}

func (la loggerAdapter) Info(msg ...any) {
	la.Logger.Info(fmt.Sprint(msg...))
}

func (la loggerAdapter) Warn(msg ...any) {
	la.Logger.Warn(fmt.Sprint(msg...))
}

func (la loggerAdapter) Error(msg ...any) {
	la.Logger.Error(fmt.Sprint(msg...))
}

func (la loggerAdapter) Debugln(msg ...any) {
	la.Logger.Sugar().Debugln(msg...)
}

func (la loggerAdapter) Infoln(msg ...any) {
	la.Logger.Sugar().Infoln(msg...)
}

func (la loggerAdapter) Warnln(msg ...any) {
	la.Logger.Sugar().Warnln(msg...)
}

func (la loggerAdapter) Errorln(msg ...any) {
	la.Logger.Sugar().Errorln(msg...)
}

func (la loggerAdapter) Debugf(format string, args ...any) {
	la.Logger.Sugar().Debugf(format, args...)

}

func (la loggerAdapter) Infof(format string, args ...any) {
	la.Logger.Sugar().Infof(format, args...)

}

func (la loggerAdapter) Warnf(format string, args ...any) {
	la.Logger.Sugar().Warnf(format, args...)
}

func (la loggerAdapter) Errorf(format string, args ...any) {
	la.Logger.Sugar().Errorf(format, args...)
}

func (la loggerAdapter) WithFields(f healthlog.Fields) healthlog.Logger {
	fields := make([]zap.Field, 0, 2*len(f))
	for name, value := range f {
		fields = append(fields, zap.Any(name, value))
	}

	return loggerAdapter{
		Logger: la.Logger.With(
			fields...,
		),
	}
}

// NewHealthLoggerAdapter adapts a go-kit logger onto the go-logger package's logger interface
func NewHealthLoggerAdapter(logger *zap.Logger) healthlog.Logger {
	return loggerAdapter{Logger: logger}
}
