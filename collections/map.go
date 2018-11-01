package collections

type Map interface {
	Get(key string) interface{}
	Put(key string, val interface{})
	Remove(key string)
	Size() int
	HasKey(key string) bool
	GetRand() interface{}
}
