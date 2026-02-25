package interfaces

import "time"

type MetricsCache interface {
	Get(key string) (any, bool)
	Set(key string, value any, duration time.Duration)
}
