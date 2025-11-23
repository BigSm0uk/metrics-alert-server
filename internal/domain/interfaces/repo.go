package interfaces

import (
	"context"

	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
)

type MetricsRepository interface {
	SaveOrUpdate(ctx context.Context, metric *domain.Metrics) error
	Metric(ctx context.Context, id, metricType string) (*domain.Metrics, error)
	MetricList(ctx context.Context) ([]domain.Metrics, error)

	SaveOrUpdateBatch(ctx context.Context, metrics []*domain.Metrics) error
	MetricListByType(ctx context.Context, metricType string) ([]domain.Metrics, error)

	Ping(ctx context.Context) error
	Close() error

	Bootstrap(ctx context.Context) error
}
