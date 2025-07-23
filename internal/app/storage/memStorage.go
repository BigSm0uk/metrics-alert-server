package storage

import (
	"sync"

	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
)

type MemStorage struct {
	db map[string]domain.Metrics
	mu sync.RWMutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{db: make(map[string]domain.Metrics)}
}

func (m *MemStorage) Set(metric domain.Metrics) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.db[metric.ID] = metric
}

func (m *MemStorage) Get(id string) (domain.Metrics, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	metric, ok := m.db[id]
	return metric, ok
}

func (m *MemStorage) GetAll() []domain.Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]domain.Metrics, 0, len(m.db))
	for _, v := range m.db {
		result = append(result, v)
	}
	return result
}

func (m *MemStorage) Delete(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.db, id)
}
