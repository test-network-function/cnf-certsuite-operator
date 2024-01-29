package controllerlogger

import (
	"fmt"

	"github.com/go-logr/logr"
)

// This struct is a wrapper for default logr, to allow the use of formatted logs.
type CollectorLogger struct {
	logr.Logger
}

func New(l logr.Logger) CollectorLogger {
	return CollectorLogger{l}
}

func (l *CollectorLogger) Infof(format string, args ...interface{}) {
	formattedMsg := fmt.Sprintf(format, args...)
	l.Info(formattedMsg)
}

func (l *CollectorLogger) Error(msg string, keysAndValues ...interface{}) {
	l.Logger.Error(nil, msg, keysAndValues...)
}

func (l *CollectorLogger) Errorf(format string, args ...interface{}) {
	formattedError := fmt.Errorf(format, args...)
	l.Error(formattedError.Error())
}
