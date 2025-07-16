package repository

import (
	"fmt"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/db"
	models "github.com/bigsm0uk/metrics-alert-server/internal/model"
)

type MetricsRepository interface {
	Save(metric *models.Metrics) error
	Get(id string) (*models.Metrics, error)
	GetAll() ([]models.Metrics, error)
	Delete(id string) error
}

type MemRepository struct {
	storage *db.MemStorage
}

func NewMemRepository(storage *db.MemStorage) *MemRepository {
	return &MemRepository{storage: storage}
}

func (r *MemRepository) Save(metric *models.Metrics) error {
	r.storage.Set(*metric)
	return nil
}

func (r *MemRepository) Get(id string) (*models.Metrics, error) {
	metric, ok := r.storage.Get(id)
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return &metric, nil
}

func (r *MemRepository) GetAll() ([]models.Metrics, error) {
	return r.storage.GetAll(), nil
}

func (r *MemRepository) Delete(id string) error {
	r.storage.Delete(id)
	return nil
}
