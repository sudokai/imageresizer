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
	for _, v := range vals {
		if v != "" {
			s.Lock()
			s.vals[v] = struct{}{}
			s.Unlock()
		}
	}
}

// Size returns the number of elements in the SyncStrSet
func (s *SyncStrSet) Size() int {
	s.RLock()
	defer s.RUnlock()
	return len(s.vals)
}

// Contains returns true if the value is present in the SyncStrSet
func (s *SyncStrSet) Contains(vals ...string) bool {
	s.RLock()
	for _, v := range vals {
		s.RUnlock()
		_, ok := s.vals[v]
		if !ok {
			return false
		}
		s.RLock()
	}
	s.RUnlock()
	return true
}

// Intersect returns the intersection of the two Sets
func (s *SyncStrSet) Intersect(s2 *SyncStrSet) *SyncStrSet {
	s.RLock()
	res := NewSyncStrSet()
	for v := range s.vals {
		s.RUnlock()
		if !s2.Contains(v) {
			s.RLock()
			continue
		}
		res.Add(v)
		s.RLock()
	}
	s.RUnlock()
	return res
}

// Walk iterates over the set, executing walkFn in a goroutine.
func (s *SyncStrSet) Walk(walkFn func(item string)) {
	s.RLock()
	defer s.RUnlock()
	for k := range s.vals {
		s.RUnlock()
		walkFn(k)
		s.RLock()
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

func (s *SyncStrSet) slice() []string {
	s.RLock()
	defer s.RUnlock()
	keys := make([]string, len(s.vals))
	i := 0
	for k := range s.vals {
		s.RUnlock()
		keys[i] = k
		i++
		s.RLock()
	}
	return keys
}
