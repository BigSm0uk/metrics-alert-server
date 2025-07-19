package service

import (
	"errors"
	"strconv"

	"go.uber.org/zap"

	models "github.com/bigsm0uk/metrics-alert-server/internal/model"
	"github.com/bigsm0uk/metrics-alert-server/internal/repository"
	"github.com/bigsm0uk/metrics-alert-server/pkg/util"
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

	m, err := s.repository.Get(id)
	if err != nil {
		m = &models.Metrics{}
	}

	switch t {
	case models.Counter:
		v, parseErr := strconv.ParseInt(value, 10, 64)
		if parseErr != nil {
			s.logger.Error("invalid counter value", zap.String("value", value), zap.Error(parseErr))
			return ErrInvalidMetricValue
		}
		nv := util.GetDefault(m.Delta) + v
		err = s.repository.Save(&models.Metrics{
			ID:    id,
			MType: t,
			Delta: &nv,
			Hash:  m.Hash,
		})
		if err != nil {
			s.logger.Error("failed to save counter metric", zap.Error(err))
			return err
		}
	case models.Gauge:
		v, parseErr := strconv.ParseFloat(value, 64)
		if parseErr != nil {
			s.logger.Error("invalid gauge value", zap.String("value", value), zap.Error(parseErr))
			return ErrInvalidMetricValue
		}
		err = s.repository.Save(&models.Metrics{
			ID:    id,
			MType: t,
			Value: &v,
			Hash:  m.Hash,
		})
		if err != nil {
			s.logger.Error("failed to save gauge metric", zap.Error(err))
			return err
		}
	}

	s.logger.Info("updating metric", zap.String("type", t), zap.String("id", id), zap.String("value", value))
	return nil
}
func (s *MetricService) GetAllMetrics() ([]models.Metrics, error) {
	m, err := s.repository.GetAll()
	if err != nil {
		return nil, err
	}
	s.logger.Info("Total metrics", zap.Int("len", len(m)))
	return m, nil
}
