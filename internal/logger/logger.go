package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

func SetupLogger(level string) error {
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	parsed, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	logrus.SetLevel(parsed)

	return nil
}
