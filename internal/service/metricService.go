package service

import (
	"errors"
	"strconv"

	"go.uber.org/zap"

	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
	"github.com/bigsm0uk/metrics-alert-server/internal/interfaces"
	"github.com/bigsm0uk/metrics-alert-server/pkg/util"
)

type MetricService struct {
	repository interfaces.MetricsRepository
	logger     *zap.Logger
}

var (
	ErrMetricNotFound     = errors.New("metric not found")
	ErrInvalidMetricType  = errors.New("invalid metric type")
	ErrInvalidMetricValue = errors.New("invalid metric value")
)

func NewService(repository interfaces.MetricsRepository, logger *zap.Logger) *MetricService {
	return &MetricService{repository: repository, logger: logger}
}
func (s *MetricService) UpdateMetric(id, t, value string) error {
	m, err := s.repository.Get(id, t)
	//Пока база в памяти реальной ошибки быть не должно
	if err != nil {
		m = &domain.Metrics{}
	}
	switch t {
	case domain.Counter:
		v, parseErr := strconv.ParseInt(value, 10, 64)
		if parseErr != nil {
			s.logger.Error("invalid counter value", zap.String("value", value), zap.Error(parseErr))
			return ErrInvalidMetricValue
		}
		nv := util.GetDefault(m.Delta) + v
		err = s.repository.Save(&domain.Metrics{
			ID:    id,
			MType: t,
			Delta: &nv,
			Hash:  m.Hash,
		})
		if err != nil {
			s.logger.Error("failed to save counter metric", zap.Error(err))
			return err
		}
	case domain.Gauge:
		v, parseErr := strconv.ParseFloat(value, 64)
		if parseErr != nil {
			s.logger.Error("invalid gauge value", zap.String("value", value), zap.Error(parseErr))
			return ErrInvalidMetricValue
		}
		err = s.repository.Save(&domain.Metrics{
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
func (s *MetricService) GetAllMetrics() ([]domain.Metrics, error) {
	m, err := s.repository.GetAll()
	if err != nil {
		return nil, err
	}
	s.logger.Debug("Total metrics", zap.Int("len", len(m)))
	return m, nil
}
func (s *MetricService) GetMetric(id, t string) (*domain.Metrics, error) {
	m, err := s.repository.Get(id, t)
	if err != nil {
		return nil, err
	}
	s.logger.Debug("Get metric", zap.String("id", id))
	return m, nil
}
