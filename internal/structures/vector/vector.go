package vector

import (
	"iter"
	"sort"

	"github.com/titosilva/drmchain-pos/internal/structures"
)

type Vector[T comparable] struct {
	Items []T
}

func New[T comparable]() *Vector[T] {
	return &Vector[T]{Items: make([]T, 0)}
}

func FromSeq[T comparable](seq iter.Seq[T]) *Vector[T] {
	v := New[T]()
	for item := range seq {
		v.Add(item)
	}

	return v
}

func (v *Vector[T]) Add(item T) {
	v.Items = append(v.Items, item)
}

func (v *Vector[T]) Remove(item T) {
	for i, it := range v.Items {
		if it == item {
			v.Items = append(v.Items[:i], v.Items[i+1:]...)
			return
		}
	}
}

func (v *Vector[T]) Contains(item T) bool {
	for _, it := range v.Items {
		if it == item {
			return true
		}
	}

	return false
}

func (v *Vector[T]) Count() int {
	return len(v.Items)
}

// All implements structures.Enumerable.
func (v *Vector[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, item := range v.Items {
			if !yield(item) {
				return
			}
		}
	}
}

// Where implements structures.Queryable.
func (v *Vector[T]) Where(predicate func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, item := range v.Items {
			if predicate(item) && !yield(item) {
				return
			}
		}
	}
}

func (v *Vector[T]) Skip(n int) structures.Enumerable[T] {
	seqFunc := func(yield func(T) bool) {
		for i, item := range v.Items {
			if i >= n && !yield(item) {
				return
			}
		}
	}

	return structures.Seq(seqFunc)
}

func (v *Vector[T]) Take(n int) structures.Enumerable[T] {
	seqFunc := func(yield func(T) bool) {
		for i, item := range v.Items {
			if i >= n {
				return
			}

			if !yield(item) {
				return
			}
		}
	}

	return structures.Seq(seqFunc)
}

// Vector implements structures.Queryable.
// Static interface impl check.
var _ structures.Queryable[int] = (*Vector[int])(nil)

// Vector implements structures.Enumerable.
// Static interface impl check.
var _ structures.Enumerable[int] = (*Vector[int])(nil)

// OrderBy implements structures.Orderable.
func (v *Vector[T]) OrderBy(isLess func(T, T) bool) iter.Seq[T] {
	items := make([]T, len(v.Items))
	copy(items, v.Items)

	sort.Slice(items, func(i, j int) bool {
		return isLess(items[i], items[j])
	})

	return func(yield func(T) bool) {
		for _, item := range items {
			if !yield(item) {
				return
			}
		}
	}
}

// Vector implements structures.Orderable.
// Static interface impl check.
var _ structures.Orderable[int] = (*Vector[int])(nil)
