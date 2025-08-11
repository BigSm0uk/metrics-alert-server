package agent

import (
	"math/rand/v2"
	"runtime"
	"sync"

	"github.com/bigsm0uk/metrics-alert-server/internal/domain"
)

type MetricsCollector struct {
	metrics   map[string]domain.Metrics
	pollCount int64
	mu        sync.RWMutex
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{metrics: make(map[string]domain.Metrics)}
}
func (c *MetricsCollector) CollectRuntimeMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	c.mu.Lock()
	defer c.mu.Unlock()

	// Gauge метрики
	gaugeMetrics := map[string]float64{
		"Alloc":         float64(m.Alloc),
		"BuckHashSys":   float64(m.BuckHashSys),
		"Frees":         float64(m.Frees),
		"GCCPUFraction": m.GCCPUFraction,
		"GCSys":         float64(m.GCSys),
		"HeapAlloc":     float64(m.HeapAlloc),
		"HeapIdle":      float64(m.HeapIdle),
		"HeapInuse":     float64(m.HeapInuse),
		"HeapObjects":   float64(m.HeapObjects),
		"HeapReleased":  float64(m.HeapReleased),
		"HeapSys":       float64(m.HeapSys),
		"LastGC":        float64(m.LastGC),
		"Lookups":       float64(m.Lookups),
		"MCacheInuse":   float64(m.MCacheInuse),
		"MCacheSys":     float64(m.MCacheSys),
		"MSpanInuse":    float64(m.MSpanInuse),
		"MSpanSys":      float64(m.MSpanSys),
		"Mallocs":       float64(m.Mallocs),
		"NextGC":        float64(m.NextGC),
		"NumForcedGC":   float64(m.NumForcedGC),
		"NumGC":         float64(m.NumGC),
		"OtherSys":      float64(m.OtherSys),
		"PauseTotalNs":  float64(m.PauseTotalNs),
		"StackInuse":    float64(m.StackInuse),
		"StackSys":      float64(m.StackSys),
		"Sys":           float64(m.Sys),
		"TotalAlloc":    float64(m.TotalAlloc),
		"RandomValue":   rand.Float64(),
	}

	for name, value := range gaugeMetrics {
		c.metrics[name] = domain.Metrics{
			ID:    name,
			MType: domain.Gauge,
			Value: &value,
		}
	}

	// Counter метрики
	c.pollCount++
	c.metrics["PollCount"] = domain.Metrics{
		ID:    "PollCount",
		MType: domain.Counter,
		Delta: &c.pollCount,
	}
}

func (c *MetricsCollector) GetMetrics() []domain.Metrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]domain.Metrics, 0, len(c.metrics))
	for _, metric := range c.metrics {
		result = append(result, metric)
	}
	return result
}
