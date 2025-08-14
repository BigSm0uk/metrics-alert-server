package repository

import (
	"context"
	"fmt"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/storage"
	"github.com/bigsm0uk/metrics-alert-server/internal/config"
	"github.com/bigsm0uk/metrics-alert-server/internal/interfaces"
)

func InitRepository(ctx context.Context, cfg *config.ServerConfig) (interfaces.MetricsRepository, error) {
	switch cfg.Storage.Type {
	case "mem":
		return NewMemRepository(storage.NewMemStorage()), nil
	case "postgres":
		if cfg.Storage.ConnectionString == "" {
			return nil, fmt.Errorf("connection string is required for postgres storage")
		}
		return NewPostgresRepository(ctx, &cfg.Storage)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", cfg.Storage.Type)
	}
}
