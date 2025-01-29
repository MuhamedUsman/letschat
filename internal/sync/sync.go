package sync

import (
	"context"
	"log/slog"
	"reflect"
	"sync"
)

// Broadcaster reads from one chan and writes to many subscribed chans, Subscribe will return a token, and a
// receive-only chan for reads, Unsubscribe must be called with the token to close the subscription,
// if not called, the system will not shut down gracefully
// Broadcast must be called in a separate long-running goroutine, this will not return even if there are no subscribers
// to relay msgs to, Broadcast will only return once shutdown is initiated
type Broadcaster[T any] struct {
	in   chan T
	mu   sync.RWMutex
	wg   sync.WaitGroup
	out  map[int]chan T
	v    T
	next int
}

func NewBroadcaster[T any]() *Broadcaster[T] {
	return &Broadcaster[T]{
		in:  make(chan T),
		wg:  sync.WaitGroup{},
		out: make(map[int]chan T),
	}
}

func (b *Broadcaster[T]) Get() T {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.v
}

func (b *Broadcaster[T]) Subscribe() (int, <-chan T) {
	c := make(chan T)
	b.mu.Lock()
	token := b.next
	b.out[token] = c
	b.next++
	b.wg.Add(1)
	b.mu.Unlock()
	return token, c
}

func (b *Broadcaster[T]) Unsubscribe(token int) {
	b.mu.Lock()
	if ch, ok := b.out[token]; ok {
		close(ch)
		delete(b.out, token)
		b.wg.Done()
	} else {
		slog.Error("channel not found while unsubscribing", "type", reflect.TypeOf(b), "token", token)
	}
	b.mu.Unlock()
}

func (b *Broadcaster[T]) Write(v T) {
	b.in <- v
}

func (b *Broadcaster[T]) Broadcast(shtdwnCtx context.Context) {
	for {
		select {
		case v := <-b.in:
			b.v = v
			b.mu.RLock() // reading from the map and writing to what we'll read, that's why RLock
			for _, ch := range b.out {
				// this may block, but we want one on one synchronization
				// if it blocks indefinitely, there is a problem elsewhere in the code
				ch <- v
			}
			b.mu.RUnlock()
		case <-shtdwnCtx.Done():
			return
		}
	}
}
