package cmap

import (
	"iter"
	"sync"

	"github.com/titosilva/drmchain-pos/internal/structures"
	"github.com/titosilva/drmchain-pos/internal/structures/kv"
)

// CMap is a concurrent map.
type CMap[K comparable, V any] struct {
	inner map[K]V // Unsafe underlying map
	mux   sync.RWMutex
}

// New creates a new concurrent map.
func New[K comparable, V any]() *CMap[K, V] {
	return &CMap[K, V]{
		inner: make(map[K]V),
		mux:   sync.RWMutex{},
	}
}

// Get retrieves a value from the map.
func (m *CMap[K, V]) Get(key K) (V, bool) {
	m.mux.RLock()
	v, ok := m.inner[key]
	m.mux.RUnlock()

	return v, ok
}

// Set sets a value in the map.
func (m *CMap[K, V]) Set(key K, value V) {
	m.mux.Lock()
	m.inner[key] = value
	m.mux.Unlock()
}

// Delete removes a value from the map.
func (m *CMap[K, V]) Delete(key K) {
	m.mux.Lock()
	delete(m.inner, key)
	m.mux.Unlock()
}

// All implements structures.Enumerable. Uses a snapshot of the map.
func (m *CMap[K, V]) All() iter.Seq[kv.KeyValue[K, V]] {
	m.mux.RLock()

	seq := make([]kv.KeyValue[K, V], 0, len(m.inner))
	for k, v := range m.inner {
		seq = append(seq, kv.KeyValue[K, V]{Key: k, Value: v})
	}

	m.mux.RUnlock()

	return func(yield func(kv.KeyValue[K, V]) bool) {
		for _, kv := range seq {
			if !yield(kv) {
				return
			}
		}
	}
}

func (m *CMap[K, V]) Skip(n int) structures.Enumerable[kv.KeyValue[K, V]] {
	all := m.All()
	return structures.Seq(all).Skip(n)
}

func (m *CMap[K, V]) Take(n int) structures.Enumerable[kv.KeyValue[K, V]] {
	all := m.All()
	return structures.Seq(all).Take(n)
}

// Where implements structures.Queryable. Uses a snapshot of the map.
func (m *CMap[K, V]) Where(predicate func(kv.KeyValue[K, V]) bool) iter.Seq[kv.KeyValue[K, V]] {
	m.mux.RLock()

	seq := make([]kv.KeyValue[K, V], 0, len(m.inner))
	for k, v := range m.inner {
		seq = append(seq, kv.KeyValue[K, V]{Key: k, Value: v})
	}

	m.mux.RUnlock()

	return func(yield func(kv.KeyValue[K, V]) bool) {
		for _, kv := range seq {
			if predicate(kv) && !yield(kv) {
				return
			}
		}
	}
}

// Count implements structures.Enumerable.
func (m *CMap[K, V]) Count() int {
	m.mux.RLock()
	count := len(m.inner)
	m.mux.RUnlock()

	return count
}

// CMap implements the Enumerable interface.
// Static interface impl check.
var _ structures.Enumerable[kv.KeyValue[int, int]] = (*CMap[int, int])(nil)

// CMap implements the Queryable interface.
// Static interface impl check.
var _ structures.Queryable[kv.KeyValue[int, int]] = (*CMap[int, int])(nil)
