package mem

import (
	"context"
	"math/rand/v2"
	"testing"

	"github.com/samber/lo"

	"github.com/bigsm0uk/metrics-alert-server/internal/app/storage"
	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
)

const (
	metricCount = 100000
)

// Тесты пока содержат только один метод, но в будущем можно добавить и другие методы для тестирования, если в них будет бизнес-логика.
// BenchmarkMemRepository_SaveOrUpdateBatch измеряет производительность сохранения/обновления пакета метрик

func BenchmarkMemRepository_SaveOrUpdateBatch(b *testing.B) {
	storage := storage.NewMemStorage()
	metrics := metricGenerator(metricCount)
	repo := NewMemRepository(storage)

	for b.Loop() {
		repo.SaveOrUpdateBatch(context.Background(), metrics)
	}
}

func metricGenerator(count uint) []*domain.Metrics {
	metrics := make([]*domain.Metrics, 0, count)
	for i := 0; i < int(count); i++ {
		metrics = append(metrics, &domain.Metrics{
			ID:    lo.RandomString(10, lo.LettersCharset),
			MType: randomChoice(domain.Gauge, domain.Counter),
			Value: lo.ToPtr(rand.Float64() * 1000),
		})
	}
	return metrics
}

func randomChoice(choices ...string) string {
	return choices[rand.IntN(len(choices))]
}
