package storage

import (
	"math/rand/v2"
	"testing"

	"github.com/samber/lo"

	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
)

const (
	// Количество метрик для подготовки данных в бенчмарках
	prepareMetricsCount = 100000
)

// BenchmarkMemStorage_Get измеряет производительность чтения одной метрики
func BenchmarkMemStorage_Get(b *testing.B) {
	storage := NewMemStorage()
	mStorage := NewMutexMemStorage()

	metrics := metricGenerator(prepareMetricsCount)
	b.Run("RWMutex", func(b *testing.B) {
		for _, metric := range metrics {
			storage.Set(*metric)
		}

		b.ResetTimer()

		for b.Loop() {
			idx := rand.IntN(len(metrics))
			storage.Get(metrics[idx].ID, metrics[idx].MType)
		}
	})
	b.Run("Mutex", func(b *testing.B) {
		for _, metric := range metrics {
			mStorage.Set(*metric)
		}

		b.ResetTimer()

		for b.Loop() {
			idx := rand.IntN(len(metrics))
			mStorage.Get(metrics[idx].ID, metrics[idx].MType)
		}
	})
}

// BenchmarkMemStorage_Set измеряет производительность записи одной метрики
func BenchmarkMemStorage_Set(b *testing.B) {
	storage := NewMemStorage()
	mStorage := NewMutexMemStorage()

	metrics := metricGenerator(prepareMetricsCount)

	b.ResetTimer()

	b.Run("RWMutex", func(b *testing.B) {
		for b.Loop() {
			idx := rand.IntN(len(metrics))
			storage.Set(*metrics[idx])
		}
	})
	b.Run("Mutex", func(b *testing.B) {
		for b.Loop() {
			idx := rand.IntN(len(metrics))
			mStorage.Set(*metrics[idx])
		}
	})
}

// BenchmarkMemStorage_GetAll измеряет производительность получения всех метрик
func BenchmarkMemStorage_GetAll(b *testing.B) {
	storage := NewMemStorage()
	mStorage := NewMutexMemStorage()
	metrics := metricGenerator(prepareMetricsCount)
	b.Run("RWMutex", func(b *testing.B) {
		for _, metric := range metrics {
			storage.Set(*metric)
		}

		b.ResetTimer()

		for b.Loop() {
			_ = storage.GetAll()
		}
	})

	b.Run("Mutex", func(b *testing.B) {
		for _, metric := range metrics {
			mStorage.Set(*metric)
		}

		b.ResetTimer()

		for b.Loop() {
			_ = mStorage.GetAll()
		}
	})
}

// BenchmarkMemStorage_GetByType измеряет производительность получения метрик по типу
func BenchmarkMemStorage_GetByType(b *testing.B) {
	storage := NewMemStorage()
	mStorage := NewMutexMemStorage()
	metrics := metricGenerator(prepareMetricsCount)
	b.Run("RWMutex", func(b *testing.B) {
		for _, metric := range metrics {
			storage.Set(*metric)
		}

		b.ResetTimer()

		for b.Loop() {
			metricType := randomChoice(domain.Counter, domain.Gauge)
			_ = storage.GetByType(metricType)
		}
	})
	b.Run("Mutex", func(b *testing.B) {
		for _, metric := range metrics {
			mStorage.Set(*metric)
		}

		b.ResetTimer()

		for b.Loop() {
			metricType := randomChoice(domain.Counter, domain.Gauge)
			_ = mStorage.GetByType(metricType)
		}
	})
}

// BenchmarkMemStorage_SetParallel измеряет производительность параллельной записи
func BenchmarkMemStorage_SetParallel(b *testing.B) {
	storage := NewMemStorage()
	mStorage := NewMutexMemStorage()
	metrics := metricGenerator(prepareMetricsCount)
	b.Run("RWMutex", func(b *testing.B) {
		for _, metric := range metrics {
			storage.Set(*metric)
		}

		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				idx := rand.IntN(len(metrics))
				storage.Set(*metrics[idx])
			}
		})
	})
	b.Run("Mutex", func(b *testing.B) {
		for _, metric := range metrics {
			mStorage.Set(*metric)
		}

		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				idx := rand.IntN(len(metrics))
				mStorage.Set(*metrics[idx])
			}
		})
	})
}

// BenchmarkMemStorage_GetParallel измеряет производительность параллельного чтения
func BenchmarkMemStorage_GetParallel(b *testing.B) {
	storage := NewMemStorage()
	mStorage := NewMutexMemStorage()
	metrics := metricGenerator(prepareMetricsCount)
	b.Run("RWMutex", func(b *testing.B) {
		for _, metric := range metrics {
			storage.Set(*metric)
		}

		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				idx := rand.IntN(len(metrics))
				storage.Get(metrics[idx].ID, metrics[idx].MType)
			}
		})
	})
	b.Run("Mutex", func(b *testing.B) {
		for _, metric := range metrics {
			mStorage.Set(*metric)
		}

		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				idx := rand.IntN(len(metrics))
				mStorage.Get(metrics[idx].ID, metrics[idx].MType)
			}
		})
	})
}

// BenchmarkMemStorage_MixedOperations измеряет смешанные операции чтения/записи
func BenchmarkMemStorage_MixedOperations(b *testing.B) {
	storage := NewMemStorage()
	mStorage := NewMutexMemStorage()
	metrics := metricGenerator(prepareMetricsCount)
	b.Run("RWMutex", func(b *testing.B) {
		for _, metric := range metrics {
			storage.Set(*metric)
		}

		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				idx := rand.IntN(len(metrics))
				if rand.IntN(100) < 80 {
					storage.Get(metrics[idx].ID, metrics[idx].MType)
				} else {
					storage.Set(*metrics[idx])
				}
			}
		})
	})
	b.Run("Mutex", func(b *testing.B) {
		for _, metric := range metrics {
			mStorage.Set(*metric)
		}

		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				idx := rand.IntN(len(metrics))
				if rand.IntN(100) < 80 {
					mStorage.Get(metrics[idx].ID, metrics[idx].MType)
				} else {
					mStorage.Set(*metrics[idx])
				}
			}
		})
	})
}

// metricGenerator генерирует случайные метрики для тестирования
func metricGenerator(count uint) []*domain.Metrics {
	metrics := make([]*domain.Metrics, count)
	for i := 0; i < int(count); i++ {
		metricType := randomChoice(domain.Gauge, domain.Counter)
		metric := &domain.Metrics{
			ID:    lo.RandomString(10, lo.LettersCharset),
			MType: metricType,
		}

		if metricType == domain.Counter {
			metric.Delta = lo.ToPtr(rand.Int64N(10000))
		} else {
			metric.Value = lo.ToPtr(rand.Float64() * 1000)
		}

		metrics[i] = metric
	}
	return metrics
}

// randomChoice возвращает случайный элемент из переданных вариантов
func randomChoice(choices ...string) string {
	return choices[rand.IntN(len(choices))]
}
