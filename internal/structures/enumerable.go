package structures

import "iter"

type Enumerable[E any] interface {
	All() iter.Seq[E]
	Count() int
	Skip(n int) Enumerable[E]
	Take(n int) Enumerable[E]
}

type SeqEnumerable[E any] struct {
	seq iter.Seq[E]
}

func (s *SeqEnumerable[E]) All() iter.Seq[E] {
	return func(yield func(E) bool) {
		for item := range s.seq {
			if !yield(item) {
				return
			}
		}
	}
}

func (s *SeqEnumerable[E]) Count() int {
	count := 0
	for range s.seq {
		count++
	}
	return count
}

func (s *SeqEnumerable[E]) Skip(n int) Enumerable[E] {
	seqFunc := func(yield func(E) bool) {
		i := 0
		for item := range s.seq {
			if i >= n && !yield(item) {
				return
			}
			i++
		}
	}

	return &SeqEnumerable[E]{seq: seqFunc}
}

func (s *SeqEnumerable[E]) Take(n int) Enumerable[E] {
	seqFunc := func(yield func(E) bool) {
		i := 0
		for item := range s.seq {
			if i >= n {
				return
			}

			if !yield(item) {
				return
			}
			i++
		}
	}

	return &SeqEnumerable[E]{seq: seqFunc}
}

func Seq[E any](seq iter.Seq[E]) Enumerable[E] {
	return &SeqEnumerable[E]{seq: seq}
}
