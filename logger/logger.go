package logger

import "sync"

type Logger interface {
	Info(format string, v ...interface{})
	Error(format string, v ...interface{})
	Fatal(format string, v ...interface{})
}

var (
	logger Logger
	mutex  = &sync.Mutex{}
)

func Info(format string, v ...interface{}) {
	logger := getLogger()
	logger.Info(format, v...)
}

func Error(format string, v ...interface{}) {
	logger := getLogger()
	logger.Error(format, v...)
}

func Fatal(format string, v ...interface{}) {
	logger := getLogger()
	logger.Fatal(format, v...)
}

func getLogger() Logger {
	mutex.Lock()
	defer mutex.Unlock()

	if logger == nil {
		logger = NewDefaultLogger()
	}

	return logger
}
