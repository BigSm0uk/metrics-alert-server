package zl

import (
	"go.uber.org/zap"
)

type Logger struct {
	*zap.Logger
}

func InitLogger(cfg zap.Config) (*zap.Logger, error) {
	return cfg.Build()
}

func InitDefaultLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}
