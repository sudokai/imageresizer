package store

type Store interface {
	Get(filename string) ([]byte, error)
	Put(filename string, buf []byte) error
	Remove(filename string) error
}

type Cache interface {
	Store
	LoadCache(walkFn func(item interface{}) error) error
	PruneCache() error
}

type Watcher interface {
	Cache
	Watch(done chan bool) error
	AddWatch(path string) error
}