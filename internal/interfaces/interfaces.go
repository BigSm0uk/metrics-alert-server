package interfaces

import "github.com/bigsm0uk/metrics-alert-server/internal/domain"

type MetricsRepository interface { //TODO добавить контекст в запросы
	Save(metric *domain.Metrics) error
	Get(id, t string) (*domain.Metrics, error)
	GetAll() ([]domain.Metrics, error)
	Delete(id string) error
}
