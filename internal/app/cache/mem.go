package cache

import (
	"sync"
	"time"
)

const (
	DefaultExpiration = 5 * time.Second
)

type Item struct {
	Value     any
	ExpiresAt time.Time
}
type Mem struct {
	items             map[string]Item
	mu                sync.RWMutex
	cleanupInterval   time.Duration
	defaultExpiration time.Duration
}

func New(defaultExpiration, cleanupInterval time.Duration) *Mem {
	items := make(map[string]Item)

	cache := Mem{
		items:             items,
		cleanupInterval:   cleanupInterval,
		defaultExpiration: defaultExpiration,
	}

	if cleanupInterval > 0 {
		cache.StartGC()
	}

	return &cache
}

func (c *Mem) StartGC() {
	go c.GC()
}

func (c *Mem) GC() {
	for {
		<-time.After(c.cleanupInterval)

		if c.items == nil {
			return
		}

		if keys := c.expiredKeys(); len(keys) != 0 {
			c.clearItems(keys)
		}
	}
}

// expiredKeys возвращает список "просроченных" ключей
func (c *Mem) expiredKeys() (keys []string) {
	c.mu.RLock()

	defer c.mu.RUnlock()

	for k, i := range c.items {
		if time.Now().After(i.ExpiresAt) {
			keys = append(keys, k)
		}
	}
	return
}

// clearItems удаляет ключи из переданного списка
func (c *Mem) clearItems(keys []string) {
	c.mu.Lock()

	defer c.mu.Unlock()

	for _, k := range keys {
		delete(c.items, k)
	}
}

func (c *Mem) Get(key string) (any, bool) {
	c.mu.RLock()

	defer c.mu.RUnlock()

	item, found := c.items[key]

	if !found {
		return nil, false
	}

	if item.ExpiresAt.IsZero() {
		if time.Now().After(item.ExpiresAt) {
			return nil, false
		}
	}

	return item.Value, true
}

func (c *Mem) Set(key string, value any, duration time.Duration) {
	var expiration int64
	if duration == 0 {
		duration = c.defaultExpiration
	}

	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}

	c.mu.Lock()

	defer c.mu.Unlock()

	c.items[key] = Item{
		Value:     value,
		ExpiresAt: time.Unix(0, expiration),
	}
}
