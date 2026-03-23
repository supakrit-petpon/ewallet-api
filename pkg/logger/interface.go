package logger

import (
	"testing"

	"go.uber.org/zap/zaptest"
)


type Logger interface {
    Info(msg string, fields ...any)
    Error(msg string, err error, fields ...any)
    Debug(msg string, fields ...any)
    Warn(msg string, fields ...any)
}

func NewTestLogger(t *testing.T) Logger {
    // zaptest.NewLogger จะผูก Log เข้ากับ t.Log ของ testing
    l := zaptest.NewLogger(t)
    return &zapLogger{log: l}
}