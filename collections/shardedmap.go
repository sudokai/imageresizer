package collections

import (
	"github.com/cespare/xxhash"
	"math/rand"
	"time"
)

type ShardedMap struct {
	shards    []*SyncMap
	shardMask uint64
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// NewShardedMap returns a map with {nShards} shards.
// {nShards} must be a power of 2
func NewShardedMap(nShards int) *ShardedMap {
	if nShards < 1 || nShards&(nShards-1) != 0 {
		// nShards must be a power of 2
		return nil
	}
	sm := &ShardedMap{
		shards:    make([]*SyncMap, nShards),
		shardMask: uint64(nShards - 1),
	}
	for i := 0; i < nShards; i++ {
		sm.shards[i] = NewSyncMap()
	}
	return sm
}

func (sm *ShardedMap) Get(key string) interface{} {
	return sm.getShard(key).Get(key)
}

func (sm *ShardedMap) Put(key string, val interface{}) {
	sm.getShard(key).Put(key, val)
}

func (sm *ShardedMap) Remove(key string) {
	sm.getShard(key).Remove(key)
}

func (sm *ShardedMap) Size() int {
	var size int
	for i := 0; i < len(sm.shards); i++ {
		size += sm.shards[i].Size()
	}
	return size
}

func (sm *ShardedMap) HasKey(key string) bool {
	return sm.getShard(key).HasKey(key)
}

func (sm *ShardedMap) GetRand() interface{} {
	x := rand.Intn(len(sm.shards))
	for i := x; i < x+len(sm.shards); i++ {
		if rnd := sm.shards[i%len(sm.shards)].GetRand(); rnd != nil {
			return rnd
		}
	}
	return nil
}

func (sm *ShardedMap) getShard(key string) *SyncMap {
	return sm.shards[xxhash.Sum64String(key)&sm.shardMask]
}
