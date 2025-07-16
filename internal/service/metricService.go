package service

import (
	"errors"

	models "github.com/bigsm0uk/metrics-alert-server/internal/model"
	"github.com/bigsm0uk/metrics-alert-server/internal/repository"
	"go.uber.org/zap"
)

type MetricService struct {
	repository repository.MetricsRepository
	logger     *zap.Logger
}

var (
	ErrInvalidMetricType  = errors.New("invalid metric type")
	ErrInvalidMetricValue = errors.New("invalid metric value")
)

func NewMetricService(repository repository.MetricsRepository, logger *zap.Logger) *MetricService {
	return &MetricService{repository: repository, logger: logger}
}
func (s *MetricService) UpdateMetric(t, id, value string) error {
	if t != models.Counter && t != models.Gauge {
		return ErrInvalidMetricType
	}

	s.logger.Info("updating metric", zap.String("type", t), zap.String("id", id), zap.String("value", value))
	return s.repository.Save(&models.Metrics{
		Type:  t,
		ID:    id,
		Value: value,
	})
}
