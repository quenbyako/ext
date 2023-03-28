// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package slices defines various functions useful with slices of any type.
// Unless otherwise specified, these functions all apply to the elements
// of a slice at index 0 <= i < len(s).
//
// Note that the less function in IsSortedFunc, SortFunc, SortStableFunc requires a
// strict weak ordering (https://en.wikipedia.org/wiki/Weak_ordering#Strict_weak_orderings),
// or the sorting may fail to sort correctly. A common case is when sorting slices of
// floating-point numbers containing NaN values.
package slices

import (
	"github.com/quenbyako/ext/constraints"
)

// Equal reports whether two slices are equal: the same length and all
// elements equal. If the lengths are different, Equal returns false.
// Otherwise, the elements are compared in increasing index order, and the
// comparison stops at the first unequal pair.
// Floating point NaNs are not considered equal.
func Equal[E comparable](s1, s2 []E) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

// EqualFunc reports whether two slices are equal using a comparison
// function on each pair of elements. If the lengths are different,
// EqualFunc returns false. Otherwise, the elements are compared in
// increasing index order, and the comparison stops at the first index
// for which eq returns false.
func EqualFunc[E1, E2 any](s1 []E1, s2 []E2, eq func(E1, E2) bool) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i, v1 := range s1 {
		v2 := s2[i]
		if !eq(v1, v2) {
			return false
		}
	}
	return true
}

// Compare compares the elements of s1 and s2.
// The elements are compared sequentially, starting at index 0,
// until one element is not equal to the other.
// The result of comparing the first non-matching elements is returned.
// If both slices are equal until one of them ends, the shorter slice is
// considered less than the longer one.
// The result is 0 if s1 == s2, -1 if s1 < s2, and +1 if s1 > s2.
// Comparisons involving floating point NaNs are ignored.
func Compare[E constraints.Ordered](s1, s2 []E) int {
	s2len := len(s2)
	for i, v1 := range s1 {
		if i >= s2len {
			return +1
		}
		v2 := s2[i]
		switch {
		case v1 < v2:
			return -1
		case v1 > v2:
			return +1
		}
	}
	if len(s1) < s2len {
		return -1
	}
	return 0
}

// CompareFunc is like Compare but uses a comparison function
// on each pair of elements. The elements are compared in increasing
// index order, and the comparisons stop after the first time cmp
// returns non-zero.
// The result is the first non-zero result of cmp; if cmp always
// returns 0 the result is 0 if len(s1) == len(s2), -1 if len(s1) < len(s2),
// and +1 if len(s1) > len(s2).
func CompareFunc[E1, E2 any](s1 []E1, s2 []E2, cmp func(E1, E2) int) int {
	s2len := len(s2)
	for i, v1 := range s1 {
		if i >= s2len {
			return +1
		}
		v2 := s2[i]
		if c := cmp(v1, v2); c != 0 {
			return c
		}
	}
	if len(s1) < s2len {
		return -1
	}
	return 0
}

func ToMap[S ~[]T, T comparable](s S) map[T]struct{} {
	res := make(map[T]struct{})
	for _, item := range s {
		res[item] = struct{}{}
	}

	return res
}

// Index returns the index of the first occurrence of v in s,
// or -1 if not present.
func Index[S ~[]T, T comparable](s S, v T) int {
	return IndexFunc(s, func(item T) bool { return item == v })
}

func IndexEq[S ~[]T, T constraints.Equal[T]](s S, v T) int {
	return IndexFunc(s, func(i T) bool { return i.Eq(v) })
}

// IndexFunc returns the first index i satisfying f(s[i]),
// or -1 if none do.
func IndexFunc[S ~[]T, T any](s S, f func(T) bool) int {
	for i, v := range s {
		if f(v) {
			return i
		}
	}
	return -1
}

func IndexLast[S ~[]T, T comparable](s S, v T) int {
	return IndexLastFunc(s, func(item T) bool { return item == v })
}

func IndexLastEq[S ~[]T, T constraints.Equal[T]](s S, v T) int {
	return IndexLastFunc(s, func(i T) bool { return i.Eq(v) })
}

func IndexLastFunc[S ~[]T, T any](s S, f func(T) bool) int {
	for i := len(s) - 1; i >= 0; i-- {
		if f(s[i]) {
			return i
		}
	}

	return -1
}

// Contains reports whether v is present in s.
func Contains[S ~[]T, T comparable](s S, v T) bool             { return Index(s, v) >= 0 }
func ContainsEq[S ~[]T, T constraints.Equal[T]](s S, v T) bool { return IndexEq(s, v) >= 0 }
func ContainsFunc[S ~[]T, T any](s S, f func(T) bool) bool     { return IndexFunc(s, f) >= 0 }

// Insert inserts the values v... into s at index i,
// returning the modified slice.
// In the returned slice r, r[i] == v[0].
// Insert panics if i is out of range.
// This function is O(len(s) + len(v)).
func Insert[S ~[]E, E any](s S, i int, v ...E) S {
	tot := len(s) + len(v)
	if tot <= cap(s) {
		s2 := s[:tot]
		copy(s2[i+len(v):], s[i:])
		copy(s2[i:], v)
		return s2
	}
	s2 := make(S, tot)
	copy(s2, s[:i])
	copy(s2[i:], v)
	copy(s2[i+len(v):], s[i:])
	return s2
}

// Delete removes the elements s[i:j] from s, returning the modified slice.
// Delete panics if s[i:j] is not a valid slice of s.
// Delete modifies the contents of the slice s; it does not create a new slice.
// Delete is O(len(s)-j), so if many items must be deleted, it is better to
// make a single call deleting them all together than to delete one at a time.
// Delete might not modify the elements s[len(s)-(j-i):len(s)]. If those
// elements contain pointers you might consider zeroing those elements so that
// objects they reference can be garbage collected.
func Delete[S ~[]E, E any](s S, i, j int) S {
	_ = s[i:j] // bounds check

	return append(s[:i], s[j:]...)
}

// Replace replaces the elements s[i:j] by the given v, and returns the
// modified slice. Replace panics if s[i:j] is not a valid slice of s.
func Replace[S ~[]E, E any](s S, i, j int, v ...E) S {
	_ = s[i:j] // verify that i:j is a valid subslice
	tot := len(s[:i]) + len(v) + len(s[j:])
	if tot <= cap(s) {
		s2 := s[:tot]
		copy(s2[i+len(v):], s[j:])
		copy(s2[i:], v)
		return s2
	}
	s2 := make(S, tot)
	copy(s2, s[:i])
	copy(s2[i:], v)
	copy(s2[i+len(v):], s[j:])
	return s2
}

// Clone returns a copy of the slice.
// The elements are copied using assignment, so this is a shallow clone.
func Clone[S ~[]E, E any](s S) S {
	if s == nil { // Preserve nil in case it matters.
		return nil
	}
	return append(S([]E{}), s...)
}

// Compact replaces consecutive runs of equal elements with a single copy.
// This is like the uniq command found on Unix.
// Compact modifies the contents of the slice s; it does not create a new slice.
func Compact[S ~[]T, T comparable](s S) S {
	return CompactFunc(s, func(a, b T) bool { return a == b })
}

func CompactEq[S ~[]T, T constraints.Equal[T]](s S) S {
	return CompactFunc(s, func(a, b T) bool { return a.Eq(b) })
}

// CompactFunc is like Compact but uses a comparison function.
func CompactFunc[S ~[]E, E any](s S, f func(E, E) bool) S {
	if len(s) < 2 {
		return s
	}
	i := 1
	last := s[0]
	for _, v := range s[1:] {
		if !f(v, last) {
			s[i] = v
			i++
			last = v
		}
	}
	return s[:i]
}

// Grow increases the slice's capacity, if necessary, to guarantee space for
// another n elements. After Grow(n), at least n elements can be appended
// to the slice without another allocation. If n is negative or too large to
// allocate the memory, Grow panics.
func Grow[S ~[]E, E any](s S, n int) S {
	if n < 0 {
		panic("cannot be negative")
	}
	if n -= cap(s) - len(s); n > 0 {
		// TODO(https://go.dev/issue/53888): Make using []E instead of S
		// to workaround a compiler bug where the runtime.growslice optimization
		// does not take effect. Revert when the compiler is fixed.
		s = append([]E(s)[:cap(s)], make([]E, n)...)[:len(s)]
	}
	return s
}

// Clip removes unused capacity from the slice, returning s[:len(s):len(s)].
func Clip[S ~[]E, E any](s S) S { return s[:len(s):len(s)] }

func Remap[S ~[]T, T, U any](s S, f func(int, T) U) []U {
	res := make([]U, len(s))
	for i, item := range s {
		res[i] = f(i, item)
	}
	return res
}

// Possibilities возвразает все возможные сочетания элементов
func Possibles[S ~[]T, T any](z []S) []S {
	if len(z) == 0 {
		return []S{}
	}
	if len(z[0]) == 0 {
		return Possibles(z[1:])
	}

	res := []S{}
	for _, elem := range z[0] {
		morePossibilities := Possibles(z[1:])
		if len(morePossibilities) == 0 {
			res = append(res, S{elem})
		}
		for _, nextItems := range morePossibilities {
			res = append(res, append(S{elem}, nextItems...))
		}
	}
	return res
}

func AppendMany[S ~[]T, T any](items ...S) S {
	res := S{}
	for _, item := range items {
		res = append(res, item...)
	}
	return res
}

func GentlyAppend[S ~[]T, T comparable](s S, items ...T) S {
	return GentlyAppendFunc(s, func(a, b T) bool { return a == b }, items...)
}

func GentlyAppendEq[S ~[]T, T constraints.Equal[T]](s S, items ...T) S {
	return GentlyAppendFunc(s, constraints.Eq[T], items...)
}

func GentlyAppendFunc[S ~[]T, T any](s S, f func(T, T) bool, items ...T) S {
	s = Grow(s, len(items))
	for _, item := range items {
		if !ContainsFunc(s, func(existed T) bool { return f(existed, item) }) {
			s = append(s, item)
		}
	}

	return Clip(s)
}

// Filter MODIFIES s, so only one possible way to use func is s = Filter(s, ...)
func Filter[S ~[]T, T any](s S, f func(T) bool) S {
	i := 0
	for _, item := range s {
		if f(item) {
			s[i] = item
			i++
		}
	}

	return Clip(s[:i])
}

func Reverse[S ~[]T, T any](s S) S {
	for i := 0; i < len(s)/2; i++ {
		j := len(s) - i - 1

		s[i], s[j] = s[j], s[i]
	}

	return s
}
