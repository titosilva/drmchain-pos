package structures

import "iter"

type Queryable[E any] interface {
	Where(predicate func(E) bool) iter.Seq[E]
}
