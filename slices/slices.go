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
	"github.com/quenbyako/ext/cmp"
)

func ToMap[S ~[]T, T comparable](s S) map[T]struct{} {
	res := make(map[T]struct{})
	for _, item := range s {
		res[item] = struct{}{}
	}

	return res
}

func IndexEq[S ~[]T, T cmp.Eq[T]](s S, v T) int {
	return IndexFunc(s, func(i T) bool { return i.Eq(v) })
}

func IndexLast[S ~[]T, T comparable](s S, v T) int {
	return IndexLastFunc(s, func(item T) bool { return item == v })
}

func IndexLastEq[S ~[]T, T cmp.Eq[T]](s S, v T) int {
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
func ContainsEq[S ~[]T, T cmp.Eq[T]](s S, v T) bool { return IndexEq(s, v) >= 0 }

func CompactEq[S ~[]T, T cmp.Eq[T]](s S) S {
	return CompactFunc(s, func(a, b T) bool { return a.Eq(b) })
}

func Remap[S ~[]E, E, T any](s S, f func(E) T) []T {
	res := make([]T, len(s))
	for i, item := range s {
		res[i] = f(item)
	}
	return res
}

func RemapIndex[S ~[]E, E, T any](s S, f func(int, E) T) []T {
	res := make([]T, len(s))
	for i, item := range s {
		res[i] = f(i, item)
	}
	return res
}

func Generate[T any](n int, f func(int) T) []T {
	res := make([]T, n)
	for i := 0; i < n; i++ {
		res[i] = f(i)
	}
	return res
}

// Possibiles возвращает все возможные сочетания элементов
func Possibles[S ~[]T, T any](s []S) (res []S) {
	if len(s) == 0 {
		return []S{}
	}
	if len(s[0]) == 0 {
		return Possibles(s[1:])
	}

	for _, elem := range s[0] {
		morePossibilities := Possibles(s[1:])
		if len(morePossibilities) == 0 {
			res = append(res, S{elem})
		}
		for _, nextItems := range morePossibilities {
			res = append(res, append(S{elem}, nextItems...))
		}
	}
	return res
}

// Concat returns a new slice concatenating the passed in slices.
func Concat[S ~[]E, E any](slices ...S) S {
	// TODO: replace for std.Concat(slices...) in 1.22
	size := 0
	for _, s := range slices {
		size += len(s)
		if size < 0 {
			panic("len out of range")
		}
	}
	newslice := Grow[S](nil, size)
	for _, s := range slices {
		newslice = append(newslice, s...)
	}
	return newslice
}

// GentlyAppend добавляет
func GentlyAppend[S ~[]T, T comparable](s S, items ...T) S {
	return GentlyAppendFunc(s, func(a, b T) bool { return a == b }, items...)
}

func GentlyAppendEq[S ~[]T, T cmp.Eq[T]](s S, items ...T) S {
	return GentlyAppendFunc(s, cmp.Equal[T], items...)
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

// AddSorted inserts items into sorted slice. This could be useful for partly
// ordered sets, but, if you need real set, use this type from other package.
func AddSorted[S ~[]T, T cmp.Ordered](s S, items ...T) S {
	return AddSortedFunc(s, cmp.Compare, items...)
}

// AddSorted inserts items of any type into sorted slice. This could be useful
// for partly ordered sets, but, if you need real set, use this type from other
// package.
func AddSortedFunc[S ~[]T, T any](s S, cmp func(elem, target T) int, items ...T) S {
	if len(items) == 0 {
		return s
	} else if len(s) == 0 {
		return SortFunc(items, cmp)
	}

	s = Grow(s, len(items))

	for _, item := range items {
		if cmp(item, s[0]) < 0 {
			s = Insert(s, 0, item)
		} else if cmp(item, s[len(s)-1]) > 0 {
			s = append(s, item)
		} else {
			i, _ := BinarySearchFunc(s, item, cmp)
			s = Insert(s, i, item)
		}
	}

	return s
}

func CountFunc[S ~[]E, E any](s S, f func(E) bool) (i int) {
	for _, t := range s {
		if f(t) {
			i++
		}
	}
	return i
}

func IsUnique[S ~[]E, E comparable](s S, v E) bool {
	if i := Index(s, v); i >= 0 {
		return Index(s[i+1:], v) < 0
	}

	return false
}

func IsUniqueFunc[S ~[]E, E any](s S, eq func(E) bool) bool {
	if i := IndexFunc(s, eq); i >= 0 {
		return IndexFunc(s[i+1:], eq) < 0
	}

	return false
}

func SortCmp[S ~[]E, E cmp.Cmp[E]](x S) S {
	return SortFunc(x, func(a, b E) int { return a.Cmp(b) })
}

func Repeat[T any](times int, s ...T) []T {
	if len(s) == 0 || times <= 0 {
		return []T{}
	}

	res := make([]T, times*len(s))
	for i := range times {
		n, m := i*len(s), (i+1)*len(s)
		copy(res[n:m], s)
	}

	return res
}
