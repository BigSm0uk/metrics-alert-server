package interfaces

import (
	"context"

	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
)

type MetricsRepository interface {
	Save(ctx context.Context, metric *domain.Metrics) error
	Get(ctx context.Context, id, metricType string) (*domain.Metrics, error)
	GetAll(ctx context.Context) ([]domain.Metrics, error)
	Delete(ctx context.Context, id string) error

	SaveBatch(ctx context.Context, metrics []domain.Metrics) error
	GetByType(ctx context.Context, metricType string) ([]domain.Metrics, error)

	Ping(ctx context.Context) error
	Close() error
}
type MetricsStore interface {
	StartProcess(ctx context.Context)
	IsActive() bool
	IsSyncMode() bool
	Close(ctx context.Context) error
	Restore(ctx context.Context) error
	SaveAllMetrics(ctx context.Context) error
	WriteMetric(metric domain.Metrics) error
}
