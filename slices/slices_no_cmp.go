//go:build !go1.22

package slices

import "cmp"

// AddSorted inserts items into sorted slice. This could be useful for partly
// ordered sets, but, if you need real set, use this type from other package.
func AddSorted[S ~[]E, E cmp.Ordered](s S, items ...E) S {

	return AddSortedFunc(s, compare[E], items...)
}

func compare[T cmp.Ordered](x, y T) int {
	switch xNaN, yNaN := x != x, y != y; {
	case xNaN && yNaN:
		return 0
	case xNaN:
		return -1
	case yNaN:
		return +1
	case x < y:
		return -1
	case x > y:
		return +1
	default:
		return 0
	}
}
