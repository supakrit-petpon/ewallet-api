// pkg/logger/zap_logger.go
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)	

type zapLogger struct {
    log *zap.Logger
}

func NewZapLogger() *zapLogger {
    config := zap.NewProductionConfig()
    config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // อ่านเวลาได้ง่ายขึ้น
    l, _ := config.Build(zap.AddCallerSkip(1))
    return &zapLogger{log: l}
}

func (z *zapLogger) Info(msg string, fields ...any) {
    // แปลง any fields เป็น zap fields (ตัวอย่างแบบย่อ)
    z.log.Info(msg) 
}

func (z *zapLogger) Error(msg string, err error, fields ...any) {
    z.log.Error(msg, zap.Error(err))
}
func (z *zapLogger) Debug(msg string, fields ...any) {
    z.log.Sugar().Debugw(msg, fields...)
}
func (z *zapLogger) Warn(msg string, fields ...any) {
    z.log.Sugar().Warnw(msg, fields...)
}