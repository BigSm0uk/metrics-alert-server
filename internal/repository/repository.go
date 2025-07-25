package repository

import (
	"fmt"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/storage"
	"github.com/bigsm0uk/metrics-alert-server/internal/config"
	"github.com/bigsm0uk/metrics-alert-server/internal/interfaces"
)

func InitRepository(cfg *config.ServerConfig) (interfaces.MetricsRepository, error) {
	switch cfg.Storage.Type {
	case "mem":
		return NewMemRepository(storage.NewMemStorage()), nil
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", cfg.Storage.Type)
	}
}
