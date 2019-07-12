package logger

import (
	"log"
	"os"
)

type defaultLogger struct {
	l *log.Logger
}

func NewDefaultLogger() Logger {
	logger = &defaultLogger{
		l: log.New(os.Stdout, "[shorturl-service] ", log.Ldate|log.Ltime),
	}
	return logger
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
