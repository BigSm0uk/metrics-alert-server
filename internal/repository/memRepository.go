package repository

import (
	"fmt"

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

func (r *MemRepository) Save(metric *domain.Metrics) error {
	r.storage.Set(*metric)
	return nil
}

func (r *MemRepository) Get(id string) (*domain.Metrics, error) {
	metric, ok := r.storage.Get(id)
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return &metric, nil
}

func (r *MemRepository) GetAll() ([]domain.Metrics, error) {
	metrics := r.storage.GetAll()
	return metrics, nil
}

func (r *MemRepository) Delete(id string) error {
	r.storage.Delete(id)
	return nil
}
