package cache

import "time"

type CacheConfig struct {
	CleanupInterval time.Duration `yaml:"cleanup_interval" env:"CACHE_CLEANUP_INTERVAL" env-default:"5s"`
}
