package log

import "go.uber.org/zap"

func NewLogger() *zap.Logger {
	logConfig := zap.NewProductionConfig()
	logConfig.Sampling = nil
	logConfig.Level.SetLevel(zap.DebugLevel)
	logger, err := logConfig.Build(zap.AddStacktrace(zap.ErrorLevel))
	if err != nil {
		panic(err)
	}
	return logger
}
