package logger

import (
	"log"
	"os"
)

// defaultLogger 默认Logger
type defaultLogger struct {
	l *log.Logger
}

// NewDefaultLogger 使用默认Logger打印日志
func NewDefaultLogger() ILogger {
	return &defaultLogger{
		l: log.New(os.Stdout, "[shorturl-service] ", log.Ldate|log.Ltime),
	}
}

func (logger *defaultLogger) Info(format string, v ...interface{}) {
	logger.l.Printf("[INFO] "+format, v...)
}

func (logger *defaultLogger) Error(format string, v ...interface{}) {
	logger.l.Printf("[ERROR] "+format, v...)
}

func (logger *defaultLogger) Fatal(format string, v ...interface{}) {
	logger.l.Printf("[ERROR] "+format, v...)
}
