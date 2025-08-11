package store

import (
	"errors"

	"github.com/bigsm0uk/metrics-alert-server/internal/config/store"
	"github.com/bigsm0uk/metrics-alert-server/internal/interfaces"
	"github.com/bigsm0uk/metrics-alert-server/internal/server/store/jsonStore"
)

func InitStore(r interfaces.MetricsRepository, cfg *store.StoreConfig) (interfaces.MetricsStore, error) {
	switch cfg.SFormat {
	case "json":
		return jsonStore.NewStore(r, cfg)
	}
	return nil, errors.New("invalid store type")
}
