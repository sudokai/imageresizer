package collections

import "testing"

func TestSetIntersect(t *testing.T) {
	s1 := NewSyncStringSet()
	s1.Add("a", "b", "c", "d")
	s2 := NewSyncStringSet()
	s2.Add("b", "c", "e", "f")
	intersection := s1.Intersect(s2)
	if !intersection.Contains("c", "b") ||
		intersection.Contains("a") ||
		intersection.Contains("d") ||
		intersection.Contains("e") ||
		intersection.Contains("f") {
		t.Errorf("Set contains wrong elements: %v", intersection.List())
	}
}