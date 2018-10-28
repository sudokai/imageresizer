package store

import "log"

type CachedStore struct {
	Store Store
	Cache Watcher
}

func (cs *CachedStore) Get(filename string) ([]byte, error) {
	var buf []byte
	var err error
	if cs.Cache != nil {
		buf, err = cs.Cache.Get(filename)
	}
	if err != nil {
		buf, err = cs.Store.Get(filename)
		if err != nil {
			return nil, err
		}
		cs.Cache.Put(filename, buf)
	}
	return buf, nil
}

func (cs *CachedStore) Put(filename string, data []byte) error {
	if cs.Cache != nil {
		err := cs.Cache.Put(filename, data)
		if err != nil {
			log.Print(err)
		}
	}
	return cs.Store.Put(filename, data)
}

func (cs *CachedStore) Remove(filename string) error {
	if cs.Cache != nil {
		err := cs.Cache.Remove(filename)
		if err != nil {
			log.Print(err)
		}
	}
	return cs.Store.Remove(filename)
}

func (cs *CachedStore) PruneCache() error {
	if cs.Cache == nil {
		return nil
	}
	return cs.Cache.PruneCache()
}

func (cs *CachedStore) LoadCache(walkFn func(item interface{}) error) error {
	if cs.Cache == nil {
		return nil
	}
	return cs.Cache.LoadCache(walkFn)
}

func (cs *CachedStore) Watch() error {
	if cs.Cache == nil {
		return nil
	}
	return cs.Cache.Watch()
}
