package repository

import (
	"context"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/storage"
	"github.com/bigsm0uk/metrics-alert-server/internal/config"
	"github.com/bigsm0uk/metrics-alert-server/internal/interfaces"
)

func InitRepository(ctx context.Context, cfg *config.ServerConfig) (interfaces.MetricsRepository, error) {
	if cfg.IsPgStoreStorage() {
		return NewPostgresRepository(ctx, &cfg.Storage)
	} else {
		return NewMemRepository(storage.NewMemStorage()), nil
	}
}
