package logger

import "sync"

type ILogger interface {
	Info(format string, v ...interface{})
	Error(format string, v ...interface{})
	Fatal(format string, v ...interface{})
}

var (
	Logger ILogger
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

func getLogger() ILogger {
	mutex.Lock()
	defer mutex.Unlock()

	if Logger == nil {
		Logger = NewDefaultLogger()
	}

	return Logger
}
