package observable

import (
	"context"
	"sync"
	"time"

	"github.com/titosilva/drmchain-pos/internal/structures/concurrency/cbag"
)

type Observable[T any] struct {
	subscribers *cbag.CBag[*Subscription[T]]
	closed      bool
	closedMux   *sync.Mutex
	bufferSize  int
}

type Subscription[T any] struct {
	observable    *Observable[T]
	channel       chan T
	closedChannel chan struct{}
	closed        bool
	closedMux     *sync.Mutex
}

func New[T any]() *Observable[T] {
	mc := new(Observable[T])

	mc.subscribers = cbag.New[*Subscription[T]]()
	mc.closedMux = &sync.Mutex{}
	mc.closed = false
	mc.bufferSize = 256

	return mc
}

func (mc *Observable[T]) Subscribe() *Subscription[T] {
	mc.closedMux.Lock()
	defer mc.closedMux.Unlock()
	if mc.closed {
		panic("channel is closed!")
	}

	s := &Subscription[T]{
		observable:    mc,
		channel:       make(chan T, mc.bufferSize),
		closed:        false,
		closedChannel: make(chan struct{}, mc.bufferSize),
		closedMux:     new(sync.Mutex),
	}

	mc.subscribers.Add(s)
	return s
}

func (mc *Observable[T]) Notify(t T) {
	mc.closedMux.Lock()
	defer mc.closedMux.Unlock()
	if mc.closed {
		return
	}

	for s := range mc.subscribers.All() {
		s.channel <- t
	}
}

func (mc *Observable[T]) unsubscribe(s *Subscription[T]) {
	mc.subscribers.Remove(s)
}

func (mc *Observable[T]) Close() {
	mc.closedMux.Lock()
	defer mc.closedMux.Unlock()
	if mc.closed {
		return
	}

	for s := range mc.subscribers.All() {
		s.Unsubscribe()
	}

	mc.closed = true
}

func (s *Subscription[T]) Unsubscribe() {
	s.Close()
	s.observable.unsubscribe(s)
}

func (s *Subscription[T]) WaitNextWithTimeoutMs(timeoutMs int) (T, bool) {
	return s.WaitNextWIthTimeout(time.Duration(timeoutMs) * time.Millisecond)
}

func (s *Subscription[T]) WaitNextWIthTimeout(timeout time.Duration) (T, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return s.WaitNextWithCancellation(ctx, cancel)
}

func (s *Subscription[T]) WaitNextWithCancellation(cancellation context.Context, cancel context.CancelFunc) (T, bool) {
	defer cancel()
	select {
	case t := <-s.channel:
		return t, true
	case <-cancellation.Done():
		return *new(T), false
	}
}

func (s *Subscription[T]) Channel() <-chan T {
	return s.channel
}

func (s *Subscription[T]) Count() int {
	return len(s.channel)
}

func (s *Subscription[T]) WaitClose() <-chan struct{} {
	return s.closedChannel
}

func (s *Subscription[T]) Close() {
	s.closedMux.Lock()
	defer s.closedMux.Unlock()
	if s.closed {
		return
	}

	close(s.channel)
	s.closedChannel <- struct{}{}
	close(s.closedChannel)
	s.closed = true
}
