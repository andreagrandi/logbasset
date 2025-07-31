package logging

import (
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func init() {
	logger = logrus.New()
	logger.SetOutput(os.Stderr)
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors:   false,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
	logger.SetLevel(logrus.InfoLevel)
}

func SetLevel(level string) error {
	logLevel, err := logrus.ParseLevel(strings.ToLower(level))
	if err != nil {
		return err
	}
	logger.SetLevel(logLevel)
	return nil
}

func SetOutput(output io.Writer) {
	logger.SetOutput(output)
}

func SetFormatter(formatter logrus.Formatter) {
	logger.SetFormatter(formatter)
}

func Debug(args ...any) {
	logger.Debug(args...)
}

func Debugf(format string, args ...any) {
	logger.Debugf(format, args...)
}

func Info(args ...any) {
	logger.Info(args...)
}

func Infof(format string, args ...any) {
	logger.Infof(format, args...)
}

func Warn(args ...any) {
	logger.Warn(args...)
}

func Warnf(format string, args ...any) {
	logger.Warnf(format, args...)
}

func Error(args ...any) {
	logger.Error(args...)
}

func Errorf(format string, args ...any) {
	logger.Errorf(format, args...)
}

func WithField(key string, value any) *logrus.Entry {
	return logger.WithField(key, value)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return logger.WithFields(fields)
}

func GetLogger() *logrus.Logger {
	return logger
}
