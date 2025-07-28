package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

func InitLogger() {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.TimeKey = "timestamp"

	var err error
	log, err = config.Build()
	if err != nil {
		panic(err)
	}
}

func L() *zap.Logger {
	return log
}

func Sync() {
	if log != nil {
		_ = log.Sync() // Flushes any buffered logs
	}
}
