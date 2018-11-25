package collections

import (
	"math"
	"testing"
)

func TestSyncMap_Get(t *testing.T) {
	sm := NewSyncMap()
	sm.Put("a", 1)
	sm.Put("b", 2)
	sm.Put("a", 3)

	if sm.Get("a") != 3 || sm.Get("b") != 2 {
		t.Errorf("Map returns wrong value")
	}

	if sm.Get("c") != nil {
		t.Errorf("Map should return nil element for non-existing key")
	}
}

func TestSyncMap_Put(t *testing.T) {
	sm := NewSyncMap()
	sm.Put("a", 1)
	sm.Put("b", 2)
	sm.Put("a", 3)

	if len(sm.indexed) != 2 {
		t.Errorf("Map's indexed slice doesn't have the correct length: %d", len(sm.indexed))
	}
	if sm.keyed["a"].idx != 0 || sm.keyed["b"].idx != 1 {
		t.Errorf("Map elements' indexes are out of sync")
	}
}

func TestSyncMap_Remove(t *testing.T) {
	sm := NewSyncMap()
	sm.Put("a", 1)
	sm.Put("b", 2)
	sm.Put("a", 3)
	sm.Remove("c")

	if len(sm.indexed) != 2 || len(sm.keyed) != 2 {
		t.Errorf("Removing non existing key actually removed an element")
	}

	sm.Remove("a")

	if len(sm.indexed) != 1 || len(sm.keyed) != 1 {
		t.Errorf("Element wasn't removed")
	}
}

func TestSyncMap_GetRand(t *testing.T) {
	sm := NewSyncMap()

	if sm.GetEvictable() != nil {
		t.Errorf("GetRand failed")
	}

	sm.Put("a", 0)
	sm.Put("b", 1)
	sm.Put("c", 2)
	sm.Put("d", 3)
	counts := []int{0, 0, 0, 0}

	for i := 0; i < 100000; i++ {
		counts[sm.GetEvictable().(int)]++
	}

	for _, x := range counts {
		if math.Abs(float64(x-25000)) > 1000 {
			t.Errorf("Not random enough")
		}
	}
}
