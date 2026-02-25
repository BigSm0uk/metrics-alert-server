package store

import (
	"errors"

	"go.uber.org/zap"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/config/store"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain/interfaces"
)

func InitStore(r interfaces.MetricsRepository, cfg *store.StoreConfig, logger *zap.Logger) (interfaces.MetricsStore, error) {
	switch cfg.SFormat {
	case "json":
		return NewJSONStore(r, cfg, logger)
	}
	return nil, errors.New("invalid store type")
}
