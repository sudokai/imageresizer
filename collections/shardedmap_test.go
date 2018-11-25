package collections

import (
	"testing"
	"time"
)

func TestNewShardedMap(t *testing.T) {
	sm := NewShardedMap(60)
	if sm != nil {
		t.Errorf("ShardedMap nShards validation failed")
	}
	sm = NewShardedMap(64)
	if sm == nil {
		t.Errorf("NewShardedMap failed")
	}
}

func TestShardedMap_Get(t *testing.T) {
	sm := NewShardedMap(64)
	sm.Put("a", 1)
	sm.Put("xyz", 2)

	if sm.Get("a") != 1 || sm.Get("xyz") != 2 {
		t.Errorf("ShardedMap.Get returned the wrong value")
	}
}

func TestShardedMap_GetRand(t *testing.T) {
	sm := NewShardedMap(64)
	ch := make(chan interface{}, 1)
	go func() {
		ch <- sm.GetEvictable()
	}()
	select {
	case <-ch:
		break
	case <-time.After(time.Second):
		t.Errorf("GetRand hang")
	}

	sm.Put("a", 1)
	count := 0
	for i := 0; i < 6400; i++ {
		if sm.GetEvictable() != nil {
			count++
		}
	}
	if count != 6400 {
		t.Errorf("GetRand returned nil value from non-empty map")
	}
}
