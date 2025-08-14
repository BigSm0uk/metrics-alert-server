package service

import (
	"context"
	"errors"
	"strconv"

	"go.uber.org/zap"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
	"github.com/bigsm0uk/metrics-alert-server/internal/interfaces"
	"github.com/bigsm0uk/metrics-alert-server/pkg/util"
)

type MetricService struct {
	repository interfaces.MetricsRepository
	store      interfaces.MetricsStore
}

var (
	ErrMetricNotFound     = errors.New("metric not found")
	ErrInvalidMetricType  = errors.New("invalid metric type")
	ErrInvalidMetricValue = errors.New("invalid metric value")
)

func NewService(repository interfaces.MetricsRepository, store interfaces.MetricsStore) *MetricService {
	return &MetricService{repository: repository, store: store}
}
func (s *MetricService) UpdateMetric(ctx context.Context, id, mType, value string) error {
	m, err := s.repository.Get(ctx, id, mType)
	//Пока база в памяти реальной ошибки быть не должно
	if err != nil {
		m = &domain.Metrics{}
	}
	switch mType {
	case domain.Counter:
		v, parseErr := strconv.ParseInt(value, 10, 64)
		if parseErr != nil {
			zl.Log.Error("invalid counter value", zap.String("value", value), zap.Error(parseErr))
			return ErrInvalidMetricValue
		}
		nv := util.GetDefault(m.Delta) + v
		err = s.repository.Save(ctx, &domain.Metrics{
			ID:    id,
			MType: mType,
			Delta: &nv,
			Hash:  m.Hash,
		})
		if err != nil {
			zl.Log.Error("failed to save counter metric", zap.Error(err))
			return err
		}
	case domain.Gauge:
		v, parseErr := strconv.ParseFloat(value, 64)
		if parseErr != nil {
			zl.Log.Error("invalid gauge value", zap.String("value", value), zap.Error(parseErr))
			return ErrInvalidMetricValue
		}
		err = s.repository.Save(ctx, &domain.Metrics{
			ID:    id,
			MType: mType,
			Value: &v,
			Hash:  m.Hash,
		})
		if err != nil {
			zl.Log.Error("failed to save gauge metric", zap.Error(err))
			return err
		}
	}

	if s.store != nil && s.store.IsSyncMode() {
		updatedMetric, _ := s.repository.Get(ctx, id, mType)
		if err := s.store.WriteMetric(*updatedMetric); err != nil {
			zl.Log.Error("failed to save metric to store", zap.Error(err))
			return err
		}
	}

	zl.Log.Debug("updating metric", zap.String("type", mType), zap.String("id", id), zap.String("value", value))
	return nil
}
func (s *MetricService) GetAllMetrics(ctx context.Context) ([]domain.Metrics, error) {
	m, err := s.repository.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	zl.Log.Debug("Total metrics", zap.Int("len", len(m)))
	return m, nil
}
func (s *MetricService) GetMetric(ctx context.Context, id, t string) (*domain.Metrics, error) {
	m, err := s.repository.Get(ctx, id, t)
	if err != nil {
		return nil, err
	}
	zl.Log.Debug("Get metric", zap.String("id", id))
	return m, nil
}
func (s *MetricService) GetEnrichMetric(ctx context.Context, id, mType string) (*domain.Metrics, error) {
	return s.repository.Get(ctx, id, mType)
}
func (s *MetricService) Ping(ctx context.Context) error {
	return s.repository.Ping(ctx)
}
