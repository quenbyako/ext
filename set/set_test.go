package set_test

import (
	"strconv"
	"testing"

	. "github.com/quenbyako/ext/set"
)

func Test_Union(t *testing.T) {
	t.Parallel()

	s := New("1", "2", "3")
	r := New("3", "4", "5")
	x := New("5", "6", "7")

	u := Union(s, r, x)
	if u.Len() != 7 {
		t.Error("Union: the merged set doesn't have all items in it.")
	}

	if !u.Has("1") || !u.Has("2") || !u.Has("3") || !u.Has("4") || !u.Has("5") || !u.Has("6") || !u.Has("7") {
		t.Error("Union: merged items are not availabile in the set.")
	}

	z := Union(x, r)
	if z.Len() != 5 {
		t.Error("Union: Union of 2 sets doesn't have the proper number of items.")
	}
}

func Test_Difference(t *testing.T) {
	t.Parallel()

	s := New("1", "2", "3")
	r := New("3", "4", "5")
	x := New("5", "6", "7")

	u := Difference(s, r, x)

	if u.Len() != 2 {
		t.Error("Difference: the set doesn't have all items in it.")
	}

	if !u.Has("1") || !u.Has("2") {
		t.Error("Difference: items are not availabile in the set.")
	}

	y := Difference(r, r)
	if y.Len() != 0 {
		t.Error("Difference: size should be zero")
	}
}

func Test_Intersection(t *testing.T) {
	t.Parallel()

	s1 := New("1", "3", "4", "5")
	s2 := New("3", "5", "6")
	s3 := New("4", "5", "6", "7")
	u := Intersection(s1, s2, s3)

	if u.Len() != 1 {
		t.Error("Intersection: the set doesn't have all items in it.")
	}

	if !u.Has("5") {
		t.Error("Intersection: items after intersection are not availabile in the set.")
	}
}

func Test_Intersection2(t *testing.T) {
	t.Parallel()

	s1 := New("1", "3", "4", "5")
	s2 := New("5", "6")
	i := Intersection(s1, s2)

	if i.Len() != 1 {
		t.Error("Intersection: size should be 1, it was", i.Len())
	}

	if !i.Has("5") {
		t.Error("Intersection: items after intersection are not availabile in the set.")
	}
}

func Test_SymmetricDifference(t *testing.T) {
	t.Parallel()

	s := New("1", "2", "3")
	r := New("3", "4", "5")
	u := SymmetricDifference(s, r)

	if u.Len() != 4 {
		t.Error("SymmetricDifference: the set doesn't have all items in it.")
	}

	if !u.Has("1") || !u.Has("2") || !u.Has("4") || !u.Has("5") {
		t.Error("SymmetricDifference: items are not availabile in the set.")
	}
}

func BenchmarkSetEquality(b *testing.B) {
	s := New[int]()
	u := New[int]()

	for i := 0; i < b.N; i++ {
		s.Add(i)
		u.Add(i)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		IsEqual(s, u)
	}
}

func BenchmarkSubset(b *testing.B) {
	s := New[int]()
	u := New[int]()

	for i := 0; i < b.N; i++ {
		s.Add(i)
		u.Add(i)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		IsSubset(s, u)
	}
}

func BenchmarkIntersection(b *testing.B) {
	for _, n := range []int{
		10,
		100,
		1000,
		10000,
		100000,
		1000000,
	} {
		s1, s2 := New[int](), New[int]()

		for i := range n / 2 {
			s1.Add(i)
		}

		for i := range n {
			s2.Add(i)
		}

		b.Run(strconv.Itoa(n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				Intersection(s1, s2)
			}
		})
	}
}
