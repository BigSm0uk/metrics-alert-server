package zl

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

func InitLogger(cfg zap.Config) (*zap.Logger, error) {
	return cfg.Build()
}

func InitDefaultLogger() *zap.Logger {
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := cfg.Build()

	return logger
}
