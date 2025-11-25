package resetter_pool

import (
	"fmt"

	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
)

// ExampleNew демонстрирует использование Pool с структурой Metrics из domain пакета
func ExampleNew() {
	// Создаем конструктор для domain.Metrics
	newFunc := func() *domain.Metrics {
		delta := int64(0)
		value := float64(0.0)
		return &domain.Metrics{
			Delta: &delta,
			Value: &value,
		}
	}

	// Создаем пул
	pool := New(newFunc)

	// Получаем объект из пула
	metrics := pool.Get()
	
	// Используем объект
	metrics.ID = "cpu_usage"
	metrics.MType = "gauge"
	*metrics.Value = 85.5
	metrics.Hash = "abc123"
	
	fmt.Printf("Используем метрику: %s = %f\n", metrics.ID, *metrics.Value)
	
	// Возвращаем объект в пул (автоматически вызовется Reset())
	pool.Put(metrics)
	
	// Получаем объект снова - он должен быть сброшен
	cleanMetrics := pool.Get()
	fmt.Printf("Сброшенная метрика: ID='%s', Value=%f\n", cleanMetrics.ID, *cleanMetrics.Value)
	
	// Output:
	// Используем метрику: cpu_usage = 85.500000
	// Сброшенная метрика: ID='', Value=0.000000
}

// ExamplePool_Get демонстрирует преимущества использования пула
func ExamplePool_Get() {
	// Создаем пул для "тяжелых" объектов
	newFunc := func() *TestStruct {
		return &TestStruct{
			Data: make([]string, 0, 1000), // предварительно выделяем большой capacity
		}
	}
	
	pool := New(newFunc)
	
	// Симулируем работу с объектами
	for i := 0; i < 5; i++ {
		obj := pool.Get()
		
		// Заполняем объект данными
		for j := 0; j < 10; j++ {
			obj.Data = append(obj.Data, fmt.Sprintf("item_%d_%d", i, j))
		}
		obj.ID = fmt.Sprintf("object_%d", i)
		obj.Value = i * 100
		
		fmt.Printf("Итерация %d: объект содержит %d элементов\n", i, len(obj.Data))
		
		// Возвращаем в пул (объект будет сброшен, но capacity сохранится)
		pool.Put(obj)
	}
	
	// Output:
	// Итерация 0: объект содержит 10 элементов
	// Итерация 1: объект содержит 10 элементов
	// Итерация 2: объект содержит 10 элементов
	// Итерация 3: объект содержит 10 элементов
	// Итерация 4: объект содержит 10 элементов
}
