package log

// TODO

import (
	"github.com/sirupsen/logrus"
)

var defaultLogger *logrus.Logger = logrus.New()

// TODO
func SetLogger(logger *logrus.Logger) {
	defaultLogger = logger
}

func Logger() *logrus.Logger {
	return defaultLogger
}
