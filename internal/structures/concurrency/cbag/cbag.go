package cbag

import (
	"iter"
	"sync"

	"github.com/titosilva/drmchain-pos/internal/structures"
)

type CBag[T comparable] struct {
	items []T
	mux   *sync.RWMutex
}

func New[T comparable]() *CBag[T] {
	return &CBag[T]{
		items: make([]T, 0),
		mux:   &sync.RWMutex{},
	}
}

func (cb *CBag[T]) Add(item T) {
	cb.mux.Lock()
	cb.items = append(cb.items, item)
	cb.mux.Unlock()
}

func (cb *CBag[T]) Remove(item T) {
	cb.mux.Lock()
	for i, v := range cb.items {
		if v == item {
			cb.items = append(cb.items[:i], cb.items[i+1:]...)
			break
		}
	}
	cb.mux.Unlock()
}

// All implements structures.Enumerable.
func (cb *CBag[T]) All() iter.Seq[T] {
	cb.mux.RLock()
	seq := make([]T, len(cb.items))
	copy(seq, cb.items)
	cb.mux.RUnlock()

	return func(yield func(T) bool) {
		for _, item := range seq {
			if !yield(item) {
				return
			}
		}
	}
}

// Count implements structures.Enumerable.
func (cb *CBag[T]) Count() int {
	cb.mux.RLock()
	count := len(cb.items)
	cb.mux.RUnlock()

	return count
}

// Skip implements structures.Enumerable.
func (cb *CBag[T]) Skip(n int) structures.Enumerable[T] {
	all := cb.All()
	return structures.Seq(all).Skip(n)
}

// Take implements structures.Enumerable.
func (cb *CBag[T]) Take(n int) structures.Enumerable[T] {
	all := cb.All()
	return structures.Seq(all).Take(n)
}

// CBag implements structures.Enumerable.
// Static implementation check
var _ structures.Enumerable[int] = (*CBag[int])(nil)
