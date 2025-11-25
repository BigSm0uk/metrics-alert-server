package resetter_pool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestStruct - тестовая структура, которая реализует интерфейс Resetter
type TestStruct struct {
	ID    string
	Value int
	Data  []string
}

// Reset сбрасывает TestStruct к начальным значениям
func (t *TestStruct) Reset() {
	t.ID = ""
	t.Value = 0
	t.Data = t.Data[:0] // очищаем слайс, но сохраняем capacity
}

func TestPool_New(t *testing.T) {
	// Создаем конструктор для TestStruct
	newFunc := func() *TestStruct {
		return &TestStruct{
			Data: make([]string, 0, 10), // предварительно выделяем capacity
		}
	}

	// Создаем пул
	pool := New(newFunc)

	assert.NotNil(t, pool)
	assert.NotNil(t, pool.new)
}

func TestPool_GetPut(t *testing.T) {
	// Создаем конструктор для TestStruct
	newFunc := func() *TestStruct {
		return &TestStruct{
			Data: make([]string, 0, 10),
		}
	}

	// Создаем пул
	pool := New(newFunc)

	// Получаем объект из пула
	obj1 := pool.Get()
	assert.NotNil(t, obj1)
	assert.Equal(t, "", obj1.ID)
	assert.Equal(t, 0, obj1.Value)
	assert.Equal(t, 0, len(obj1.Data))

	// Изменяем объект
	obj1.ID = "test-id"
	obj1.Value = 42
	obj1.Data = append(obj1.Data, "item1", "item2")

	// Возвращаем объект в пул
	pool.Put(obj1)

	// Получаем объект снова (может быть тот же самый)
	obj2 := pool.Get()
	assert.NotNil(t, obj2)

	// Проверяем, что объект был сброшен
	assert.Equal(t, "", obj2.ID)
	assert.Equal(t, 0, obj2.Value)
	assert.Equal(t, 0, len(obj2.Data))
	// Capacity должен сохраниться
	assert.Equal(t, 10, cap(obj2.Data))
}

func TestPool_MultipleObjects(t *testing.T) {
	// Создаем конструктор для TestStruct
	newFunc := func() *TestStruct {
		return &TestStruct{
			Data: make([]string, 0, 5),
		}
	}

	// Создаем пул
	pool := New(newFunc)

	// Получаем несколько объектов
	obj1 := pool.Get()
	obj2 := pool.Get()
	obj3 := pool.Get()

	// Изменяем их
	obj1.ID = "obj1"
	obj1.Value = 1

	obj2.ID = "obj2"
	obj2.Value = 2

	obj3.ID = "obj3"
	obj3.Value = 3

	// Возвращаем в пул
	pool.Put(obj1)
	pool.Put(obj2)
	pool.Put(obj3)

	// Получаем объекты снова
	newObj1 := pool.Get()
	newObj2 := pool.Get()
	newObj3 := pool.Get()

	// Все объекты должны быть сброшены
	assert.Equal(t, "", newObj1.ID)
	assert.Equal(t, 0, newObj1.Value)

	assert.Equal(t, "", newObj2.ID)
	assert.Equal(t, 0, newObj2.Value)

	assert.Equal(t, "", newObj3.ID)
	assert.Equal(t, 0, newObj3.Value)
}

// TestMetrics - тестовая структура, аналогичная Metrics из domain
type TestMetrics struct {
	ID    string
	MType string
	Delta *int64
	Value *float64
	Hash  string
}

// Reset сбрасывает TestMetrics к начальным значениям
func (m *TestMetrics) Reset() {
	m.ID = ""
	m.MType = ""
	if m.Delta != nil {
		*m.Delta = 0
	}
	if m.Value != nil {
		*m.Value = 0.0
	}
	m.Hash = ""
}

// Пример использования с структурой, аналогичной Metrics из domain
func TestPool_WithMetrics(t *testing.T) {
	// Создаем конструктор
	newFunc := func() *TestMetrics {
		delta := int64(0)
		value := float64(0.0)
		return &TestMetrics{
			Delta: &delta,
			Value: &value,
		}
	}

	// Создаем пул
	pool := New(newFunc)

	// Тестируем
	metrics := pool.Get()
	assert.NotNil(t, metrics)
	assert.Equal(t, "", metrics.ID)
	assert.Equal(t, "", metrics.MType)
	assert.Equal(t, int64(0), *metrics.Delta)
	assert.Equal(t, float64(0.0), *metrics.Value)

	// Изменяем значения
	metrics.ID = "test-metric"
	metrics.MType = "counter"
	*metrics.Delta = 100
	*metrics.Value = 3.14
	metrics.Hash = "test-hash"

	// Возвращаем в пул
	pool.Put(metrics)

	// Получаем снова
	newMetrics := pool.Get()
	assert.Equal(t, "", newMetrics.ID)
	assert.Equal(t, "", newMetrics.MType)
	assert.Equal(t, int64(0), *newMetrics.Delta)
	assert.Equal(t, float64(0.0), *newMetrics.Value)
	assert.Equal(t, "", newMetrics.Hash)
}
