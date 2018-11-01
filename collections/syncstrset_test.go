package collections

import "testing"

func TestSyncStrSet_Intersect(t *testing.T) {
	s1 := NewSyncStrSet()
	s1.Add("a", "b", "c", "d")
	s2 := NewSyncStrSet()
	s2.Add("b", "c", "e", "f")
	intersection := s1.Intersect(s2)
	if !intersection.Contains("c", "b") ||
		intersection.Contains("a") ||
		intersection.Contains("d") ||
		intersection.Contains("e") ||
		intersection.Contains("f") {
		t.Errorf("Set contains wrong elements: %v", intersection.Slice())
	}
}

func TestSyncStrSet_Slice(t *testing.T) {
	set := NewSyncStrSet()
	set.Add("a", "b", "c", "d")
	slice := set.slice()
	for _, x := range []string{"a", "b", "c", "d"} {
		var found bool
		for _, v := range slice {
			found = x == v
			if found {
				break
			}
		}
		if !found {
			t.Errorf("Slice doesn't contain all the set's elements")
		}
	}
	if len(slice) != 4 {
		t.Errorf("Wrong slice length")
	}
}
