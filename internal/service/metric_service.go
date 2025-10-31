package service

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/zl"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain/interfaces"
	"github.com/bigsm0uk/metrics-alert-server/internal/repository/strategy"
	"github.com/bigsm0uk/metrics-alert-server/pkg/util"
)

type MetricService struct {
	repository interfaces.MetricsRepository
	store      interfaces.MetricsStore
}

func NewService(repository interfaces.MetricsRepository, store interfaces.MetricsStore) *MetricService {
	return &MetricService{repository: repository, store: store}
}

func (s *MetricService) SaveOrUpdateMetric(ctx context.Context, metric *domain.Metrics) error {
	// Получаем существующую метрику или создаем пустую для новой
	oldMetric, err := s.repository.Metric(ctx, metric.ID, metric.MType)
	if err != nil {
		if errors.Is(err, domain.ErrMetricNotFound) {
			zl.Log.Debug("new metric", zap.String("id", metric.ID), zap.String("type", metric.MType))
			// Для новой метрики создаем пустую с нулевыми значениями
			oldMetric = &domain.Metrics{
				ID:    metric.ID,
				MType: metric.MType,
			}
		} else {
			return err
		}
	}

	// Получаем стратегию обновления для типа метрики
	updateStrategy := strategy.StrategyFactory(metric.MType)
	if updateStrategy == nil {
		return fmt.Errorf("unsupported metric type: %s", metric.MType)
	}

	// Применяем стратегию обновления
	updatedMetric := updateStrategy.Update(oldMetric, metric)

	// Сохраняем обновленную метрику
	err = s.repository.SaveOrUpdate(ctx, updatedMetric)
	if err != nil {
		zl.Log.Error("failed to save metric",
			zap.Error(err),
			zap.String("type", metric.MType),
			zap.String("id", metric.ID),
		)
		return err
	}

	// Синхронизация с хранилищем, если требуется
	if s.store != nil && s.store.IsActive() && s.store.IsSyncMode() {
		if err := s.store.WriteMetric(*updatedMetric); err != nil {
			zl.Log.Error("failed to save metric to store", zap.Error(err))
			return err
		}
	}

	zl.Log.Debug("updating metric",
		zap.String("type", metric.MType),
		zap.String("id", metric.ID),
		zap.String("value", fmt.Sprintf("%v", util.GetDefault(metric.Value))),
		zap.String("delta", fmt.Sprintf("%v", util.GetDefault(metric.Delta))),
	)
	return nil
}

func (s *MetricService) SaveOrUpdateMetricsBatch(ctx context.Context, metrics []*domain.Metrics) error {
	err := s.repository.SaveOrUpdateBatch(ctx, metrics)
	if err != nil {
		zl.Log.Error("failed to save metrics batch", zap.Error(err))
		return err
	}
	if s.store != nil && s.store.IsActive() && s.store.IsSyncMode() {
		s.store.SaveAllMetrics(ctx)
	}
	return nil
}

func (s *MetricService) GetAllMetrics(ctx context.Context) ([]domain.Metrics, error) {
	m, err := s.repository.MetricList(ctx)
	if err != nil {
		return nil, err
	}
	zl.Log.Debug("Total metrics", zap.Int("len", len(m)))
	return m, nil
}

func (s *MetricService) GetMetric(ctx context.Context, id, t string) (*domain.Metrics, error) {
	m, err := s.repository.Metric(ctx, id, t)
	if err != nil {
		return nil, err
	}
	zl.Log.Debug("Get metric", zap.String("id", id))
	return m, nil
}

func (s *MetricService) GetEnrichMetric(ctx context.Context, id, mType string) (*domain.Metrics, error) {
	return s.repository.Metric(ctx, id, mType)
}

func (s *MetricService) Ping(ctx context.Context) error {
	return s.repository.Ping(ctx)
}

func (s *MetricService) Close() error {
	return s.repository.Close()
}
