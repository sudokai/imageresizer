package collections

import "sync"

type SyncMap struct {
	data map[string]interface{}
	sync.Mutex
}

func NewSyncMap() *SyncMap {
	sm := &SyncMap{
		data: map[string]interface{}{},
	}
	return sm
}

func (sm *SyncMap) Get(key string) interface{} {
	sm.Lock()
	defer sm.Unlock()
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
	sm.Lock()
	defer sm.Unlock()
	return len(sm.data)
}

func (sm *SyncMap) HasKey(key string) bool {
	sm.Lock()
	defer sm.Unlock()
	_, ok := sm.data[key]
	return ok
}

func (sm *SyncMap) GetRand() interface{} {
	for _, v := range sm.data {
		return v
	}
	return nil
}
