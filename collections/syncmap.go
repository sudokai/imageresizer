package collections

import "sync"

type SyncMap struct {
	data map[string]interface{}
	sync.RWMutex
}

func NewSyncMap() *SyncMap {
	sm := &SyncMap{
		data: map[string]interface{}{},
	}
	return sm
}

func (sm *SyncMap) Get(key string) interface{} {
	sm.RLock()
	defer sm.RUnlock()
	return sm.data[key]
}

func (sm *SyncMap) Put(key string, val interface{}) {
	sm.Lock()
	defer sm.Unlock()
	sm.data[key] = val
}

func (sm *SyncMap) Remove(key string) {
	sm.Lock()
	defer sm.Unlock()
	delete(sm.data, key)
}

func (sm *SyncMap) Size() int {
	sm.RLock()
	defer sm.RUnlock()
	return len(sm.data)
}

func (sm *SyncMap) HasKey(key string) bool {
	sm.RLock()
	defer sm.RUnlock()
	_, ok := sm.data[key]
	return ok
}

func (sm *SyncMap) GetRand() interface{} {
	sm.RLock()
	defer sm.RUnlock()
	for _, v := range sm.data {
		return v
	}
	return nil
}
