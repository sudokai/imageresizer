package store

type NoopCache struct{}

func (c *NoopCache) Get(filename string) ([]byte, error) {
	return nil, nil
}
func (c *NoopCache) Put(filename string, buf []byte) error {
	return nil
}
func (c *NoopCache) Remove(filename string) error {
	return nil
}
func (c *NoopCache) LoadCache(walkFn func(item interface{}) error) error {
	return nil
}
func (c *NoopCache) PruneCache() error {
	return nil
}
