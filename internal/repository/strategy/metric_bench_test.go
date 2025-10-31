package strategy

import (
	"testing"

	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
	"github.com/samber/lo"
)

// BenchmarkStrategyFactory_Singleton тестирует производительность с
func BenchmarkStrategyFactory_Singleton(b *testing.B) {
	oldMetric := &domain.Metrics{
		ID:    "test",
		MType: domain.Counter,
		Delta: lo.ToPtr(int64(100)),
	}

	newMetric := &domain.Metrics{
		ID:    "test",
		MType: domain.Counter,
		Delta: lo.ToPtr(int64(50)),
	}

	for b.Loop() {
		strategy := StrategyFactory(domain.Counter)
		_ = strategy.Update(oldMetric, newMetric)
	}
}

// BenchmarkStrategyFactory_WithAllocation демонстрирует что было бы с аллокацией
func BenchmarkStrategyFactory_WithAllocation(b *testing.B) {
	oldMetric := &domain.Metrics{
		ID:    "test",
		MType: domain.Counter,
		Delta: lo.ToPtr(int64(100)),
	}

	newMetric := &domain.Metrics{
		ID:    "test",
		MType: domain.Counter,
		Delta: lo.ToPtr(int64(50)),
	}

	for b.Loop() {
		// Создаем новый экземпляр каждый раз (старая версия)
		strategy := &CounterStrategy{}
		_ = strategy.Update(oldMetric, newMetric)
	}
}

// BenchmarkCounterStrategy_Update тестирует производительность обновления counter
func BenchmarkCounterStrategy_Update(b *testing.B) {
	strategy := counterStrategy

	oldMetric := &domain.Metrics{
		ID:    "requests",
		MType: domain.Counter,
		Delta: lo.ToPtr(int64(1000)),
		Hash:  "old_hash",
	}

	newMetric := &domain.Metrics{
		ID:    "requests",
		MType: domain.Counter,
		Delta: lo.ToPtr(int64(100)),
		Hash:  "new_hash",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strategy.Update(oldMetric, newMetric)
	}
}

// BenchmarkGaugeStrategy_Update тестирует производительность обновления gauge
func BenchmarkGaugeStrategy_Update(b *testing.B) {
	strategy := gaugeStrategy

	oldMetric := &domain.Metrics{
		ID:    "cpu_usage",
		MType: domain.Gauge,
		Value: lo.ToPtr(75.5),
		Hash:  "old_hash",
	}

	newMetric := &domain.Metrics{
		ID:    "cpu_usage",
		MType: domain.Gauge,
		Value: lo.ToPtr(82.3),
		Hash:  "new_hash",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strategy.Update(oldMetric, newMetric)
	}
}

// BenchmarkParallelStrategyFactory тестирует производительность при параллельном доступе
func BenchmarkParallelStrategyFactory(b *testing.B) {
	oldMetric := &domain.Metrics{
		ID:    "test",
		MType: domain.Counter,
		Delta: lo.ToPtr(int64(100)),
	}

	newMetric := &domain.Metrics{
		ID:    "test",
		MType: domain.Counter,
		Delta: lo.ToPtr(int64(50)),
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			strategy := StrategyFactory(domain.Counter)
			_ = strategy.Update(oldMetric, newMetric)
		}
	})
}

// BenchmarkMixedStrategies тестирует смешанное использование разных стратегий
func BenchmarkMixedStrategies(b *testing.B) {
	counterOld := &domain.Metrics{
		ID:    "counter",
		MType: domain.Counter,
		Delta: lo.ToPtr(int64(100)),
	}
	counterNew := &domain.Metrics{
		ID:    "counter",
		MType: domain.Counter,
		Delta: lo.ToPtr(int64(50)),
	}

	gaugeOld := &domain.Metrics{
		ID:    "gauge",
		MType: domain.Gauge,
		Value: lo.ToPtr(75.5),
	}
	gaugeNew := &domain.Metrics{
		ID:    "gauge",
		MType: domain.Gauge,
		Value: lo.ToPtr(82.3),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Попеременно используем разные стратегии
		if i%2 == 0 {
			strategy := StrategyFactory(domain.Counter)
			_ = strategy.Update(counterOld, counterNew)
		} else {
			strategy := StrategyFactory(domain.Gauge)
			_ = strategy.Update(gaugeOld, gaugeNew)
		}
	}
}
