package cqueue

import (
	"sync"

	"github.com/titosilva/drmchain-pos/internal/structures/queue"
)

type CQueue[T any] struct {
	underlying *queue.Queue[T]
	mux        *sync.Mutex
}

func New[T any]() *CQueue[T] {
	return &CQueue[T]{
		underlying: queue.New[T](),
		mux:        &sync.Mutex{},
	}
}

func (cq *CQueue[T]) Enqueue(data T) *CQueue[T] {
	cq.mux.Lock()
	defer cq.mux.Unlock()

	cq.underlying.Enqueue(data)
	return cq
}

func (cq *CQueue[T]) Dequeue() (T, bool) {
	cq.mux.Lock()
	defer cq.mux.Unlock()

	return cq.underlying.Dequeue()
}
