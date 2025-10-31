package strategy

import (
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
	"github.com/samber/lo"
)

// MetricUpdateStrategy определяет интерфейс для обновления метрик.
// Принимает старую метрику и новую метрику с обновленными данными.
type MetricUpdateStrategy interface {
	Update(oldMetric, newMetric *domain.Metrics) *domain.Metrics
}

// CounterStrategy реализует логику обновления counter метрик (суммирование).
type CounterStrategy struct{}

func (s *CounterStrategy) Update(oldMetric, newMetric *domain.Metrics) *domain.Metrics {
	// Counter метрики суммируются
	oldDelta := lo.FromPtr(oldMetric.Delta)
	newDelta := lo.FromPtr(newMetric.Delta)

	return &domain.Metrics{
		ID:    newMetric.ID,
		MType: domain.Counter,
		Delta: lo.ToPtr(oldDelta + newDelta),
		Hash:  newMetric.Hash,
	}
}

// GaugeStrategy реализует логику обновления gauge метрик (замена значения).
type GaugeStrategy struct{}

func (s *GaugeStrategy) Update(oldMetric, newMetric *domain.Metrics) *domain.Metrics {
	// Gauge метрики заменяют старое значение новым
	return &domain.Metrics{
		ID:    newMetric.ID,
		MType: domain.Gauge,
		Value: newMetric.Value,
		Hash:  newMetric.Hash,
	}
}

// Singleton экземпляры стратегий - создаются один раз при инициализации пакета.
// Поскольку стратегии не имеют состояния (stateless), они безопасны для многопоточного использования.
var (
	counterStrategy = &CounterStrategy{}
	gaugeStrategy   = &GaugeStrategy{}
)

// StrategyFactory возвращает соответствующую стратегию для типа метрики.
// Возвращает singleton экземпляр вместо создания нового объекта каждый раз.
func StrategyFactory(metricType string) MetricUpdateStrategy {
	switch metricType {
	case domain.Counter:
		return counterStrategy
	case domain.Gauge:
		return gaugeStrategy
	default:
		return nil
	}
}
