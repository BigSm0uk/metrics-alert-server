package repository

import (
	"context"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/config"
	"github.com/bigsm0uk/metrics-alert-server/internal/app/storage"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain/interfaces"
	"github.com/bigsm0uk/metrics-alert-server/internal/repository/mem"
	"github.com/bigsm0uk/metrics-alert-server/internal/repository/pg"
)

func InitRepository(ctx context.Context, cfg *config.ServerConfig) (interfaces.MetricsRepository, error) {
	if cfg.IsPgStoreStorage() {
		return pg.NewPostgresRepository(ctx, &cfg.Storage)
	} else {
		return mem.NewMemRepository(storage.NewMemStorage()), nil
	}
}
