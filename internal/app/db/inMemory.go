package db

import "sync"

type InMemory struct {
	db map[string]interface{}
	mu sync.RWMutex
}

func NewInMemory() *InMemory {
	return &InMemory{db: make(map[string]any)}
}
