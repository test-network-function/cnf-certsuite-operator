package controllerlogger

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// This struct is a wrapper for default logr, to allow the use of formatted logs.
type ControllerLogger struct {
	logr.Logger
}

func New() ControllerLogger {
	return ControllerLogger{Logger: log.FromContext(context.TODO())}
}

func (l *ControllerLogger) Infof(format string, args ...interface{}) {
	formattedMsg := fmt.Sprintf(format, args...)
	l.Info(formattedMsg)
}

func (l *ControllerLogger) Error(msg string, keysAndValues ...interface{}) {
	l.Logger.Error(nil, msg, keysAndValues...)
}

func (l *ControllerLogger) Errorf(format string, args ...interface{}) {
	formattedError := fmt.Errorf(format, args...)
	l.Error(formattedError.Error())
}
