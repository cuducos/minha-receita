package transform

type semaphore struct {
	semaphore chan struct{}
}

func (s *semaphore) Get() {
	s.semaphore <- struct{}{}
}

func (s *semaphore) Release() {
	<-s.semaphore
}

func newSemaphore(n int) *semaphore {
	s := semaphore{make(chan struct{}, n)}
	return &s
}
