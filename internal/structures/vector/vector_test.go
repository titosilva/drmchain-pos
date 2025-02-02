package vector_test

import (
	"testing"

	"github.com/titosilva/drmchain-pos/internal/structures/vector"
)

func Test__Take__ShouldLimitTheQuantityOfItemsReturned(t *testing.T) {
	vec := vector.New[int]()
	vec.Add(1)
	vec.Add(2)
	vec.Add(3)
	vec.Add(4)
	vec.Add(5)

	taken := vec.Take(3)
	for item := range taken.All() {
		if item > 3 {
			t.Errorf("Take() should limit the quantity of items returned")
		}
	}
}

func Test__Skip__ShouldSkipTheFirstItems(t *testing.T) {
	vec := vector.New[int]()
	vec.Add(1)
	vec.Add(2)
	vec.Add(3)
	vec.Add(4)
	vec.Add(5)

	skipped := vec.Skip(3)
	for item := range skipped.All() {
		if item < 4 {
			t.Errorf("Skip() should skip the first items")
		}
	}
}
