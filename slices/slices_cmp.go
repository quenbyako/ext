//go:build go1.22

package slices

import "cmp"

// AddSorted inserts items into sorted slice. This could be useful for partly
// ordered sets, but, if you need real set, use this type from other package.
func AddSorted[S ~[]T, T cmp.Ordered](s S, items ...T) S {
	return AddSortedFunc(s, cmp.Compare, items...)
}
