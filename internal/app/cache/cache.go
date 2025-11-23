package cache

import (
	"time"

	"github.com/bigsm0uk/metrics-alert-server/internal/domain/interfaces"
)

// NewCache инициализирует кеш с заданным интервалом очистки (в будущем можно расширить)
func NewCache(cleanupInterval time.Duration) interfaces.MetricsCache {
	return New(DefaultExpiration, cleanupInterval)
}
