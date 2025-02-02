package clru_test

import (
	"testing"

	"github.com/titosilva/drmchain-pos/internal/structures/concurrency/clru"
)

func Test__ClruCacheShouldEvictWhenCacheGetsFull(t *testing.T) {
	cache := clru.New[int, int](2)

	cache.Put(1, 1)
	cache.Put(2, 2)
	cache.Put(3, 3)

	if _, ok := cache.Get(1); ok {
		t.Error("Expected 1 to be evicted")
	}

	cache.Get(2)
	cache.Put(4, 4)

	if _, ok := cache.Get(3); ok {
		t.Error("Expected 3 to be evicted")
	}
}
