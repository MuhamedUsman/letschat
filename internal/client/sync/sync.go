package sync

import (
	"context"
	"sync"
)

type StateMonitor[T any] struct {
	c    chan T
	cond *sync.Cond
	s    T
}

func NewStatus[T any](initial T) *StateMonitor[T] {
	return &StateMonitor[T]{
		c:    make(chan T),
		cond: sync.NewCond(new(sync.Mutex)),
		s:    initial,
	}
}

// Get will return the underlying conditioned variable, keep in mind it will be affected by race conditions
func (s *StateMonitor[T]) Get() T {
	return s.s
}

// GetAndBlock will return the underlying conditioned variable, keep in mind you must call Unblock after you're done
func (s *StateMonitor[T]) GetAndBlock() T {
	s.cond.L.Lock()
	return s.s
}

func (s *StateMonitor[T]) Unblock() {
	s.cond.L.Unlock()
}

// WaitForStateChange will block until there is a state change broadcast internally
func (s *StateMonitor[T]) WaitForStateChange() T {
	s.cond.L.Lock()
	s.cond.Wait()
	defer s.cond.L.Unlock()
	return s.s
}

// WriteToChan writes the passed value to the chan, Broadcast must be called afterward
func (s *StateMonitor[T]) WriteToChan(v T) {
	s.c <- v
}

func (s *StateMonitor[T]) Broadcast(shtdwnCtx context.Context) {
	for {
		select {
		case val := <-s.c:
			s.cond.L.Lock()
			s.s = val
			s.cond.L.Unlock()
			s.cond.Broadcast()
		case <-shtdwnCtx.Done():
			// all the goroutines that waits for the broadcast also respects the shtdwnCtx so after the wait
			// they first check the shtdwnCtx if that's done they will exit
			s.cond.Broadcast()
			return
		}
	}
}
