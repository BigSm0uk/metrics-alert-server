package strategy

import (
	"testing"

	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestCounterStrategy_Update(t *testing.T) {
	tests := []struct {
		name      string
		oldMetric *domain.Metrics
		newMetric *domain.Metrics
		expected  *domain.Metrics
	}{
		{
			name: "добавление к существующему counter",
			oldMetric: &domain.Metrics{
				ID:    "requests",
				MType: domain.Counter,
				Delta: lo.ToPtr(int64(100)),
				Hash:  "old_hash",
			},
			newMetric: &domain.Metrics{
				ID:    "requests",
				MType: domain.Counter,
				Delta: lo.ToPtr(int64(50)),
				Hash:  "new_hash",
			},
			expected: &domain.Metrics{
				ID:    "requests",
				MType: domain.Counter,
				Delta: lo.ToPtr(int64(150)), // 100 + 50
				Hash:  "new_hash",
			},
		},
		{
			name: "новый counter с нулевым старым значением",
			oldMetric: &domain.Metrics{
				ID:    "new_counter",
				MType: domain.Counter,
				Delta: lo.ToPtr(int64(0)),
			},
			newMetric: &domain.Metrics{
				ID:    "new_counter",
				MType: domain.Counter,
				Delta: lo.ToPtr(int64(10)),
				Hash:  "hash",
			},
			expected: &domain.Metrics{
				ID:    "new_counter",
				MType: domain.Counter,
				Delta: lo.ToPtr(int64(10)),
				Hash:  "hash",
			},
		},
		{
			name: "отрицательные значения counter",
			oldMetric: &domain.Metrics{
				ID:    "errors",
				MType: domain.Counter,
				Delta: lo.ToPtr(int64(5)),
			},
			newMetric: &domain.Metrics{
				ID:    "errors",
				MType: domain.Counter,
				Delta: lo.ToPtr(int64(-3)),
			},
			expected: &domain.Metrics{
				ID:    "errors",
				MType: domain.Counter,
				Delta: lo.ToPtr(int64(2)), // 5 + (-3)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := &CounterStrategy{}
			result := strategy.Update(tt.oldMetric, tt.newMetric)

			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.MType, result.MType)
			assert.Equal(t, *tt.expected.Delta, *result.Delta)
			if tt.expected.Hash != "" {
				assert.Equal(t, tt.expected.Hash, result.Hash)
			}
		})
	}
}

func TestGaugeStrategy_Update(t *testing.T) {
	tests := []struct {
		name      string
		oldMetric *domain.Metrics
		newMetric *domain.Metrics
		expected  *domain.Metrics
	}{
		{
			name: "замена значения gauge",
			oldMetric: &domain.Metrics{
				ID:    "cpu_usage",
				MType: domain.Gauge,
				Value: lo.ToPtr(75.5),
				Hash:  "old_hash",
			},
			newMetric: &domain.Metrics{
				ID:    "cpu_usage",
				MType: domain.Gauge,
				Value: lo.ToPtr(82.3),
				Hash:  "new_hash",
			},
			expected: &domain.Metrics{
				ID:    "cpu_usage",
				MType: domain.Gauge,
				Value: lo.ToPtr(82.3), // полностью заменяется
				Hash:  "new_hash",
			},
		},
		{
			name: "обнуление gauge",
			oldMetric: &domain.Metrics{
				ID:    "temp",
				MType: domain.Gauge,
				Value: lo.ToPtr(100.0),
			},
			newMetric: &domain.Metrics{
				ID:    "temp",
				MType: domain.Gauge,
				Value: lo.ToPtr(0.0),
			},
			expected: &domain.Metrics{
				ID:    "temp",
				MType: domain.Gauge,
				Value: lo.ToPtr(0.0),
			},
		},
		{
			name: "отрицательное значение gauge",
			oldMetric: &domain.Metrics{
				ID:    "balance",
				MType: domain.Gauge,
				Value: lo.ToPtr(50.0),
			},
			newMetric: &domain.Metrics{
				ID:    "balance",
				MType: domain.Gauge,
				Value: lo.ToPtr(-25.5),
			},
			expected: &domain.Metrics{
				ID:    "balance",
				MType: domain.Gauge,
				Value: lo.ToPtr(-25.5),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := &GaugeStrategy{}
			result := strategy.Update(tt.oldMetric, tt.newMetric)

			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.MType, result.MType)
			assert.Equal(t, *tt.expected.Value, *result.Value)
			if tt.expected.Hash != "" {
				assert.Equal(t, tt.expected.Hash, result.Hash)
			}
		})
	}
}

func TestStrategyFactory(t *testing.T) {
	tests := []struct {
		name         string
		metricType   string
		expectNil    bool
		strategyType string
	}{
		{
			name:         "counter strategy",
			metricType:   domain.Counter,
			expectNil:    false,
			strategyType: "*strategy.CounterStrategy",
		},
		{
			name:         "gauge strategy",
			metricType:   domain.Gauge,
			expectNil:    false,
			strategyType: "*strategy.GaugeStrategy",
		},
		{
			name:       "неизвестный тип метрики",
			metricType: "unknown",
			expectNil:  true,
		},
		{
			name:       "пустой тип метрики",
			metricType: "",
			expectNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := StrategyFactory(tt.metricType)

			if tt.expectNil {
				assert.Nil(t, strategy)
			} else {
				assert.NotNil(t, strategy)
				// Проверяем, что возвращается правильная стратегия
				switch tt.metricType {
				case domain.Counter:
					_, ok := strategy.(*CounterStrategy)
					assert.True(t, ok, "ожидается CounterStrategy")
				case domain.Gauge:
					_, ok := strategy.(*GaugeStrategy)
					assert.True(t, ok, "ожидается GaugeStrategy")
				}
			}
		})
	}
}

func TestStrategyPattern_Integration(t *testing.T) {
	t.Run("полный цикл обновления метрик с паттерном стратегия", func(t *testing.T) {
		// Исходное состояние метрик
		oldCounterMetric := &domain.Metrics{
			ID:    "total_requests",
			MType: domain.Counter,
			Delta: lo.ToPtr(int64(1000)),
			Hash:  "hash1",
		}

		oldGaugeMetric := &domain.Metrics{
			ID:    "memory_usage",
			MType: domain.Gauge,
			Value: lo.ToPtr(512.5),
			Hash:  "hash2",
		}

		// Новые значения для обновления
		newCounterMetric := &domain.Metrics{
			ID:    "total_requests",
			MType: domain.Counter,
			Delta: lo.ToPtr(int64(250)),
			Hash:  "hash3",
		}

		newGaugeMetric := &domain.Metrics{
			ID:    "memory_usage",
			MType: domain.Gauge,
			Value: lo.ToPtr(768.0),
			Hash:  "hash4",
		}

		// Применяем стратегии
		counterStrategy := StrategyFactory(domain.Counter)
		gaugeStrategy := StrategyFactory(domain.Gauge)

		updatedCounter := counterStrategy.Update(oldCounterMetric, newCounterMetric)
		updatedGauge := gaugeStrategy.Update(oldGaugeMetric, newGaugeMetric)

		// Проверяем результаты
		assert.Equal(t, int64(1250), *updatedCounter.Delta, "counter должен суммироваться")
		assert.Equal(t, "hash3", updatedCounter.Hash)

		assert.Equal(t, 768.0, *updatedGauge.Value, "gauge должен заменяться")
		assert.Equal(t, "hash4", updatedGauge.Hash)
	})
}
