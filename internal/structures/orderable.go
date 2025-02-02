package structures

import "iter"

type Orderable[T any] interface {
	Enumerable[T]
	OrderBy(isLess func(T, T) bool) iter.Seq[T]
}
