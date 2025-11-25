package resetpool

import "github.com/bigsm0uk/metrics-alert-server/internal/domain"

// MetricsPool создает пул для структуры domain.Metrics
func MetricsPool() *Pool[*domain.Metrics] {
	return New(func() *domain.Metrics {
		delta := int64(0)
		value := float64(0.0)
		return &domain.Metrics{
			Delta: &delta,
			Value: &value,
		}
	})
}

// Пример использования:
//
//	pool := util.MetricsPool()
//
//	// Получаем метрику из пула
//	metric := pool.Get()
//
//	// Используем метрику
//	metric.ID = "cpu_usage"
//	metric.MType = "gauge"
//	*metric.Value = 85.5
//
//	// Возвращаем в пул (автоматически сбросится)
//	pool.Put(metric)
