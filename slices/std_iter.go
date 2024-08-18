//go:build go1.23

package slices

import (
	"cmp"
	"iter"
	std "slices"
)

// All returns an iterator over index-value pairs in the slice in the usual
// order.
func All[Slice ~[]E, E any](s Slice) iter.Seq2[int, E] { return std.All(s) }

// AppendSeq appends the values from seq to the slice and returns the extended
// slice.
func AppendSeq[Slice ~[]E, E any](s Slice, seq iter.Seq[E]) Slice { return std.AppendSeq(s, seq) }

// Backward returns an iterator over index-value pairs in the slice, traversing
// it backward with descending indices.
func Backward[Slice ~[]E, E any](s Slice) iter.Seq2[int, E] { return std.Backward(s) }

// Chunk returns an iterator over consecutive sub-slices of up to n elements of
// s. All but the last sub-slice will have size n. All sub-slices are clipped to
// have no capacity beyond the length. If s is empty, the sequence is empty:
// there is no empty slice in the sequence. Chunk panics if n is less than 1.
func Chunk[Slice ~[]E, E any](s Slice, n int) iter.Seq[Slice] { return std.Chunk(s, n) }

// Collect collects values from seq into a new slice and returns it.
func Collect[E any](seq iter.Seq[E]) []E { return std.Collect(seq) }

// Repeat returns a new slice that repeats the provided slice the given number
// of times. The result has length and capacity (len(x) * count). The result is
// never nil. Repeat panics if count is negative or if the result of (len(x) *
// count) overflows.
func Repeat[S ~[]E, E any](x S, count int) S { return std.Repeat(x, count) }

// Sorted collects values from seq into a new slice, sorts the slice, and
// returns it.
func Sorted[E cmp.Ordered](seq iter.Seq[E]) []E { return std.Sorted(seq) }

// SortedFunc collects values from seq into a new slice, sorts the slice using
// the comparison function, and returns it.
func SortedFunc[E any](seq iter.Seq[E], cmp func(E, E) int) []E { return std.SortedFunc(seq, cmp) }

// SortedStableFunc collects values from seq into a new slice. It then sorts the
// slice while keeping the original order of equal elements, using the
// comparison function to compare elements. It returns the new slice.
func SortedStableFunc[E any](seq iter.Seq[E], cmp func(E, E) int) []E {
	return std.SortedStableFunc(seq, cmp)
}

// Values returns an iterator that yields the slice elements in order.
func Values[Slice ~[]E, E any](s Slice) iter.Seq[E] { return std.Values(s) }
