package zl

import (
	"go.uber.org/zap"
)

type Logger struct {
	zl *zap.Logger
}

func InitLoggerMust(cfg zap.Config) *zap.Logger {
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return logger
}
