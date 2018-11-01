package store

type Cache interface {
	Store
	LoadCache(walkFn func(item interface{}) error) error
	PruneCache() error
}