package collections

import "sync"

// SyncStrSet is a collection of unique values
type SyncStrSet struct {
	vals map[string]struct{}
	sync.RWMutex
}

// NewSyncStrSet returns a new SyncStrSet
func NewSyncStrSet() *SyncStrSet {
	s := &SyncStrSet{
		vals: make(map[string]struct{}),
	}
	return s
}

// Add adds a value to the SyncStrSet
func (s *SyncStrSet) Add(vals ...string) {
	s.Lock()
	defer s.Unlock()
	for _, v := range vals {
		if v != "" {
			s.vals[v] = struct{}{}
		}
	}
}

// Size returns the number of elements in the SyncStrSet
func (s *SyncStrSet) Size() int {
	s.RLock()
	defer s.RUnlock()
	return len(s.vals)
}

// List returns the values of the SyncStrSet as a slice
func (s *SyncStrSet) Slice() []string {
	s.RLock()
	defer s.RUnlock()
	keys := make([]string, len(s.vals))
	i := 0
	for k := range s.vals {
		keys[i] = k
		i++
	}
	return keys
}

// Contains returns true if the value is present in the SyncStrSet
func (s *SyncStrSet) Contains(vals ...string) bool {
	s.RLock()
	defer s.RUnlock()
	for _, v := range vals {
		_, ok := s.vals[v]
		if !ok {
			return false
		}
	}
	return true
}

// Intersect returns the intersection of the two Sets
func (s *SyncStrSet) Intersect(s2 *SyncStrSet) *SyncStrSet {
	s.RLock()
	defer s.RUnlock()
	res := NewSyncStrSet()
	for v := range s.vals {
		if !s2.Contains(v) {
			continue
		}
		res.Add(v)
	}
	return res
}

// Walk iterates over the set, executing walkFn in a goroutine.
func (s *SyncStrSet) Walk(walkFn func(item string)) {
	s.RLock()
	defer s.RUnlock()
	for k := range s.vals {
		go walkFn(k)
	}
}

// Get returns a random element of the set
func (s *SyncStrSet) Get() string {
	s.RLock()
	defer s.RUnlock()
	for k := range s.vals {
		return k
	}
	return ""
}

// Remove removes element x from the set
func (s *SyncStrSet) Remove(x string) {
	s.Lock()
	defer s.Unlock()
	if x != "" {
		delete(s.vals, x)
	}
}