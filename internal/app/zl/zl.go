package zl

import (
	"go.uber.org/zap"
)

type Logger struct {
	zl *zap.Logger
}

func InitLogger(cfg zap.Config) (*zap.Logger, error) {
	return cfg.Build()
}
