package collections

import "sync"

// SyncStrSet is a collection of unique values
type SyncStrSet struct {
	vals map[string]struct{}
	sync.Mutex
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
		s.vals[v] = struct{}{}
	}
}

// Size returns the number of elements in the SyncStrSet
func (s *SyncStrSet) Size() int {
	s.Lock()
	defer s.Unlock()
	return len(s.vals)
}

// List returns the values of the SyncStrSet as a slice
func (s *SyncStrSet) List() []string {
	s.Lock()
	defer s.Unlock()
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
	s.Lock()
	defer s.Unlock()
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
	s.Lock()
	defer s.Unlock()
	res := NewSyncStrSet()
	for v := range s.vals {
		if !s2.Contains(v) {
			continue
		}
		res.Add(v)
	}
	return res
}

// Walk iterates over the set, executing walkFn.
// If walkFn returns a function, interrupt the walk and return the error
func (s *SyncStrSet) Walk(walkFn func(item string) error) error {
	s.Lock()
	defer s.Unlock()
	for k := range s.vals {
		err := walkFn(k)
		if err != nil {
			return err
		}
	}
	return nil
}