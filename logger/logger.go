package logger

import (
	"github.com/sirupsen/logrus"
)

func New(config Config) *Logger {
	logger := &Logger{
		Logger: *(logrus.New()),
		config: config,
	}
	logger.Level = logrus.WarnLevel
	// log.Formatter = &logrus.TextFormatter{}

	return logger
}

type Logger struct {
	logrus.Logger
	config Config
}

func (l *Logger) addHook(hook logrus.Hook) {
	l.Logger.Hooks.Add(hook)
}

type Config struct {
}
