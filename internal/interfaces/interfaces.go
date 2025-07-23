package interfaces

import "github.com/bigsm0uk/metrics-alert-server/internal/domain"

type MetricsRepository interface {
	Save(metric *domain.Metrics) error
	Get(id string) (*domain.Metrics, error)
	GetAll() ([]domain.Metrics, error)
	Delete(id string) error
}
