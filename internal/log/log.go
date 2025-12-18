package log

import (
	"github.com/sirupsen/logrus"
)

// Init initialize logger
func Init(level string) {
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})

	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logrus.WithError(err).Warnf("Invalid log level '%s', defaulting to 'info'", level)
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(logLevel)
	}

	logrus.Info("Logger initialized")
}
