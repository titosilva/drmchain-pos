package structures

import "iter"

func Map[S, R any](proj func(S) R, src Enumerable[S]) Enumerable[R] {
	return Seq(MapIter(proj, src.All()))
}

func MapIter[S, R any](proj func(S) R, src iter.Seq[S]) iter.Seq[R] {
	return func(yield func(R) bool) {
		for item := range src {
			if !yield(proj(item)) {
				return
			}
		}
	}
}
