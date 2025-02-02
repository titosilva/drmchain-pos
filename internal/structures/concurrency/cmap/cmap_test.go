package cmap_test

import (
	"testing"

	"github.com/titosilva/drmchain-pos/internal/structures/concurrency/cmap"
	"github.com/titosilva/drmchain-pos/internal/structures/kv"
)

func Test__CMapShouldReturnCorrectValues__WhenFilled(t *testing.T) {
	// Arrange
	cm := cmap.New[int, int]()

	// Act
	cm.Set(1, 2)
	cm.Set(2, 3)
	cm.Set(3, 4)

	// Assert
	if v, ok := cm.Get(1); !ok || v != 2 {
		t.Error("Could not get value 2 for key 1")
	}
}

func Test__CMapShouldReturnAllValues__WhenCallingAll(t *testing.T) {
	// Arrange
	cm := cmap.New[int, int]()

	// Act
	cm.Set(1, 2)
	cm.Set(2, 3)
	cm.Set(3, 4)

	// Assert
	seq := cm.All()
	count := 0

	for kv := range seq {
		count++
		if kv.Value != kv.Key+1 {
			t.Errorf("Unexpected key-value pair: %v", kv)
		}
	}

	if count != 3 {
		t.Error("Unexpected count of key-value pairs")
	}
}

func Test__CMapShouldReturnCorrectValues__WhenDeleted(t *testing.T) {
	// Arrange
	cm := cmap.New[int, int]()

	// Act
	cm.Set(1, 2)
	cm.Set(2, 3)
	cm.Set(3, 4)

	// Assert
	cm.Delete(2)

	if _, ok := cm.Get(2); ok {
		t.Error("Key 2 should have been deleted")
	}
}

func Test__CMapShouldReturnCorrectValues__WhenFiltered(t *testing.T) {
	// Arrange
	cm := cmap.New[int, int]()

	// Act
	cm.Set(1, 2)
	cm.Set(2, 3)
	cm.Set(3, 4)

	// Assert
	seq := cm.Where(func(kv kv.KeyValue[int, int]) bool {
		return kv.Key%2 == 0
	})

	count := 0
	for kv := range seq {
		count++
		if kv.Key%2 != 0 {
			t.Errorf("Unexpected key-value pair: %v", kv)
		}
	}

	if count != 1 {
		t.Error("Unexpected count of key-value pairs")
	}
}
