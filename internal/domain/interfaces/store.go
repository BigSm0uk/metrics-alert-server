package interfaces

import (
	"context"

	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
)

type MetricsStore interface {
	StartProcess(ctx context.Context)
	IsActive() bool
	IsSyncMode() bool
	Close(ctx context.Context) error
	Restore(ctx context.Context) error
	SaveAllMetrics(ctx context.Context) error
	WriteMetric(metric domain.Metrics) error
}
