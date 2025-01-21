package set_test

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/quenbyako/ext/set"
)

func Test_New(t *testing.T) {
	t.Parallel()

	s := New("1", "2", "3", "testing")

	if s.Len() != 4 {
		t.Error("New: The set created was expected have 4 items")
	}
}

func TestSetNonTS_Add(t *testing.T) {
	t.Parallel()

	s := New[string]()
	s.Add("1")
	s.Add("2")
	s.Add("2") // duplicate
	s.Add("fatih")
	s.Add("zeynep")
	s.Add("zeynep") // another duplicate

	if s.Len() != 4 {
		t.Error("Add: items are not unique. The set size should be four")
	}

	if !s.Has("1") || !s.Has("2") || !s.Has("fatih") || !s.Has("zeynep") {
		t.Error("Add: added items are not availabile in the set.")
	}
}

func TestSetNonTS_Add_multiple(t *testing.T) {
	t.Parallel()

	s := New("ankara", "san francisco", "3.14")

	if s.Len() != 3 {
		t.Error("Add: items are not unique. The set size should be three")
	}

	if !s.Has("ankara") || !s.Has("san francisco") || !s.Has("3.14") {
		t.Error("Add: added items are not availabile in the set.")
	}
}

func TestSetNonTS_Remove(t *testing.T) {
	t.Parallel()

	s := New(1, 2, 3)

	s.Del(1)

	if s.Len() != 2 {
		t.Error("Remove: set size should be two after removing")
	}

	s.Del(1)

	if s.Len() != 2 {
		t.Error("Remove: set size should be not change after trying to remove a non-existing item")
	}

	s.Del(2)
	s.Del(3)

	if s.Len() != 0 {
		t.Error("Remove: set size should be zero")
	}

	s.Del(3) // try to remove something from a zero length set
}

func TestSetNonTS_Remove_multiple(t *testing.T) {
	t.Parallel()

	s := New("ankara", "san francisco", "3.14", "istanbul")

	s.Del("ankara", "san francisco", "3.14")

	if s.Len() != 1 {
		t.Error("Remove: items are not unique. The set size should be four")
	}

	if !s.Has("istanbul") {
		t.Error("Add: added items are not availabile in the set.")
	}
}

func TestSetNonTS_Pop(t *testing.T) {
	t.Parallel()

	s := New[int]()
	s.Add(1)
	s.Add(2)
	s.Add(3)

	s, a, ok := Pop(s)
	if !ok {
		t.Error("Pop: expected to get value")
	}

	if s.Len() != 2 {
		t.Error("Pop: set size should be two after popping out")
	}

	if s.Has(a) {
		t.Error("Pop: returned item should not exist")
	}

	s, _, _ = Pop(s)
	s, _, _ = Pop(s)

	if s, _, ok = Pop(s); ok {
		t.Error("Pop: expected to not get value")
	}
}

func TestSetNonTS_Has(t *testing.T) {
	t.Parallel()

	s := New("1", "2", "3", "4")

	if !s.Has("1") {
		t.Error("Has: the item 1 exist, but 'Has' is returning false")
	}

	if !s.Has("1") || !s.Has("2") || !s.Has("3") || !s.Has("4") {
		t.Error("Has: the items all exist, but 'Has' is returning false")
	}
}

func TestSetNonTS_IsEqual(t *testing.T) {
	t.Parallel()

	s := New("1", "2", "3")
	u := New("1", "2", "3")

	if ok := IsEqual(s, u); !ok {
		t.Error("IsEqual: set s and t are equal. However it returns false")
	}

	// same size, different content
	a := New("1", "2", "3")
	b := New("4", "5", "6")

	if ok := IsEqual(a, b); ok {
		t.Error("IsEqual: set a and b are now equal (1). However it returns true")
	}

	// different size, similar content
	a = New("1", "2", "3")
	b = New("1", "2", "3", "4")

	if ok := IsEqual(a, b); ok {
		t.Error("IsEqual: set s and t are now equal (2). However it returns true")
	}
}

func TestSetNonTS_IsSubset(t *testing.T) {
	t.Parallel()

	s := New("1", "2", "3", "4")
	u := New("1", "2", "3")

	if ok := IsSubset(s, u); !ok {
		t.Error("IsSubset: u is a subset of s. However it returns false")
	}

	if ok := IsSubset(u, s); ok {
		t.Error("IsSubset: s is not a subset of u. However it returns true")
	}
}

func TestSetNonTS_IsSuperset(t *testing.T) {
	t.Parallel()

	s := New("1", "2", "3", "4")
	u := New("1", "2", "3")

	if ok := IsSuperset(u, s); !ok {
		t.Error("IsSuperset: s is a superset of u. However it returns false")
	}

	if ok := IsSuperset(s, u); ok {
		t.Error("IsSuperset: u is not a superset of u. However it returns true")
	}
}

func TestSetNonTS_String(t *testing.T) {
	t.Parallel()

	s := New[string]()
	if str := fmt.Sprintf("%v", s); str != "set[]" {
		t.Errorf("String: output is not what is excepted '%s'", str)
	}

	s.Add("1", "2", "3", "4")

	str := fmt.Sprintf("%v", s)

	if !strings.HasPrefix(str, "set[") {
		t.Error("String: output should begin with a square bracket")
	}

	if !strings.HasSuffix(str, "]") {
		t.Error("String: output should end with a square bracket")
	}
}

func TestSetNonTS_Copy(t *testing.T) {
	t.Parallel()

	s := New("1", "2", "3", "4")
	r := Clone(s)

	if !IsEqual(s, r) {
		t.Error("Copy: set s and r are not equal")
	}
}

func TestSetNonTS_Subtract(t *testing.T) {
	t.Parallel()

	s := New("1", "2", "3")
	r := New("3", "5")
	s = Subtract(s, r)

	if s.Len() != 2 {
		t.Error("Subtract: the set doesn't have all items in it.")
	}

	if !s.Has("1") || !s.Has("2") {
		t.Error("Subtract: items after separation are not availabile in the set.")
	}
}
