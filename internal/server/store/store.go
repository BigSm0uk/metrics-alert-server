package store

import (
	"errors"

	"github.com/bigsm0uk/metrics-alert-server/internal/config/store"
	"github.com/bigsm0uk/metrics-alert-server/internal/interfaces"
)

func InitStore(r interfaces.MetricsRepository, cfg *store.StoreConfig) (interfaces.MetricsStore, error) {
	switch cfg.SFormat {
	case "json":
		return NewJsonStore(r, cfg)
	}
	return nil, errors.New("invalid store type")
}
