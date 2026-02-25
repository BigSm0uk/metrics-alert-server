package storage

import (
	"sync"

	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
)

// MutexMemStorage - версия хранилища с обычным Mutex (вместо RWMutex).
// Блокирует на чтение и запись одинаково.
type MutexMemStorage struct {
	db map[string]domain.Metrics
	mu sync.Mutex
}

// NewMutexMemStorage создает новый экземпляр хранилища
func NewMutexMemStorage() *MutexMemStorage {
	return &MutexMemStorage{db: make(map[string]domain.Metrics)}
}

// Set добавляет/обновляет метрику
func (m *MutexMemStorage) Set(metric domain.Metrics) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.db[metric.ID] = metric
}

// Get получает метрику по ID и типу
func (m *MutexMemStorage) Get(id, t string) (domain.Metrics, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	metric, ok := m.db[id]
	if !ok {
		return domain.Metrics{}, false
	}
	if metric.MType != t {
		return domain.Metrics{}, false
	}
	return metric, ok
}

// GetByType получает все метрики определенного типа
func (m *MutexMemStorage) GetByType(metricType string) []domain.Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]domain.Metrics, 0, len(m.db))
	for _, v := range m.db {
		if v.MType == metricType {
			result = append(result, v)
		}
	}
	return result
}

// GetAll возвращает все метрики
func (m *MutexMemStorage) GetAll() []domain.Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]domain.Metrics, 0, len(m.db))
	for _, v := range m.db {
		result = append(result, v)
	}
	return result
}

// Delete удаляет метрику по ID
func (m *MutexMemStorage) Delete(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.db, id)
}
