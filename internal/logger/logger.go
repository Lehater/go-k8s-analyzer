package logger

import (
    "log"
)

type Logger struct {
    level string
}

func New(level string) *Logger {
    if level == "" {
        level = "info"
    }
    return &Logger{level: level}
}

func (l *Logger) Info(msg string, args ...interface{}) {
    log.Printf("[INFO] "+msg, args...)
}

func (l *Logger) Error(msg string, args ...interface{}) {
    log.Printf("[ERROR] "+msg, args...)
}

func (l *Logger) Debug(msg string, args ...interface{}) {
    if l.level == "debug" {
        log.Printf("[DEBUG] "+msg, args...)
    }
}
