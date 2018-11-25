package collections

type EvictableMap interface {
	Get(key string) interface{}
	Put(key string, val interface{})
	Remove(key string)
	HasKey(key string) bool
	GetEvictable() interface{}
}
