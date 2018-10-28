package collections

import (
	"math/rand"
	"sync"
	"time"
)

type item struct {
	idx int
	key string
	val interface{}
}

type SyncMap struct {
	indexed []*item
	keyed   map[string]*item
	sync.RWMutex
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewSyncMap() *SyncMap {
	sm := &SyncMap{
		keyed: map[string]*item{},
	}
	return sm
}

func (sm *SyncMap) Get(key string) interface{} {
	sm.RLock()
	defer sm.RUnlock()
	elem, ok := sm.keyed[key]
	if !ok {
		return nil
	}
	return elem.val
}

func (sm *SyncMap) Put(key string, val interface{}) {
	sm.Lock()
	defer sm.Unlock()
	if elem, ok := sm.keyed[key]; ok {
		elem.val = val
		return
	}
	elem := &item{idx: len(sm.indexed), key: key, val: val}
	sm.keyed[key] = elem
	sm.indexed = append(sm.indexed, elem)
}

func (sm *SyncMap) Remove(key string) {
	sm.Lock()
	defer sm.Unlock()
	elem, ok := sm.keyed[key]
	if !ok {
		return
	}

	delete(sm.keyed, key)

	lastIdx := len(sm.indexed) - 1
	if elem.idx < lastIdx {
		// elem.idx is not the last
		// swap the element we delete with the last element
		last := sm.indexed[lastIdx]
		last.idx = elem.idx
		sm.indexed[elem.idx] = last
	}
	// delete the last element
	sm.indexed[lastIdx] = nil // https://github.com/golang/go/wiki/SliceTricks
	sm.indexed = sm.indexed[:lastIdx]
}

func (sm *SyncMap) Size() int {
	sm.RLock()
	defer sm.RUnlock()
	return len(sm.indexed)
}

func (sm *SyncMap) HasKey(key string) bool {
	sm.RLock()
	defer sm.RUnlock()
	_, ok := sm.keyed[key]
	return ok
}

func (sm *SyncMap) GetRand() interface{} {
	sm.RLock()
	defer sm.RUnlock()
	elem := sm.indexed[rand.Intn(len(sm.indexed))]
	return sm.Get(elem.key)
}
