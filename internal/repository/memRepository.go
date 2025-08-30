package repository

import (
	"context"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/storage"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
	"github.com/bigsm0uk/metrics-alert-server/internal/interfaces"
)

type MemRepository struct {
	storage *storage.MemStorage
}

var _ interfaces.MetricsRepository = (*MemRepository)(nil)

func NewMemRepository(storage *storage.MemStorage) *MemRepository {
	return &MemRepository{storage: storage}
}

func (r *MemRepository) SaveOrUpdate(ctx context.Context, metric *domain.Metrics) error {
	r.storage.Set(*metric)
	return nil
}

func (r *MemRepository) Get(ctx context.Context, id, t string) (*domain.Metrics, error) {
	metric, ok := r.storage.Get(id, t)
	if !ok {
		return nil, domain.ErrMetricNotFound
	}
	return &metric, nil
}

func (r *MemRepository) GetAll(ctx context.Context) ([]domain.Metrics, error) {
	metrics := r.storage.GetAll()
	return metrics, nil
}

func (r *MemRepository) SaveOrUpdateBatch(ctx context.Context, metrics []domain.Metrics) error {
	for _, metric := range metrics {
		r.storage.Set(metric)
	}
	return nil
}
func (r *MemRepository) GetByType(ctx context.Context, metricType string) ([]domain.Metrics, error) {
	metrics := r.storage.GetByType(metricType)
	return metrics, nil
}
func (r *MemRepository) Ping(ctx context.Context) error {
	return nil
}
func (r *MemRepository) Close() error {
	return nil
}
func (r *MemRepository) MustBootstrap(ctx context.Context) {
	return
}
