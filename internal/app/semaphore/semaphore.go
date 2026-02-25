package semaphore

type Semaphore struct {
	semaphore chan struct{}
}

func NewSemaphore(size int) *Semaphore {
	return &Semaphore{semaphore: make(chan struct{}, size)}
}

func (s *Semaphore) Acquire() {
	s.semaphore <- struct{}{}
}

func (s *Semaphore) Release() {
	<-s.semaphore
}
