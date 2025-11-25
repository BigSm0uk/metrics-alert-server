package resetter_pool

import "sync"

// Resetter определяет интерфейс для объектов, которые могут быть сброшены к начальному состоянию
type Resetter interface {
	Reset()
}

// Pool представляет собой пул объектов с generic-параметром, ограниченным типами с методом Reset()
type Pool[T Resetter] struct {
	pool sync.Pool
	new  func() T
}

// New создает и возвращает указатель на новую структуру Pool
func New[T Resetter](newFunc func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() interface{} {
				return newFunc()
			},
		},
		new: newFunc,
	}
}

// Get возвращает объект из пула
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put помещает объект в пул после вызова его метода Reset()
func (p *Pool[T]) Put(obj T) {
	// Важно: сбрасываем состояние объекта перед возвратом в пул
	obj.Reset()
	p.pool.Put(obj)
}
