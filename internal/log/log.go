package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger() *zap.Logger {
	logConfig := zap.NewProductionConfig()
	logConfig.Sampling = nil
	logConfig.Level.SetLevel(zap.DebugLevel)
	logConfig.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	logger, err := logConfig.Build(zap.AddStacktrace(zap.ErrorLevel))
	if err != nil {
		panic(err)
	}
	return logger
}
