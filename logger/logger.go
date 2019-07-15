package logger

import "sync"

// ILogger 实现该接口可以自定义日志打印
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

// getLogger 获取Logger，如果Logger为nil会使用defaultLogger
func getLogger() ILogger {
	mutex.Lock()
	defer mutex.Unlock()

	if Logger == nil {
		Logger = NewDefaultLogger()
	}

	return Logger
}
