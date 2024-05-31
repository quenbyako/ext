// Copyright (c) 2020-2024 Richard Cooper
//
// This file is a part of quenbyako/ext package.
// See https://github.com/quenbyako/ext/blob/master/LICENSE for details

package span

import (
	"cmp"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/quenbyako/ext/slices"
)

// обозначения для документации:
//
// n, m — верхняя и нижняя границы баунда
//
// [n:m] — включительно
// (n:m) — не включительно
// [n:m) — включительно слева, не включительно справа
// {n:m} — включение не определено (не имеет значения для описываемого действия)
// {n:m] — включение не определено слева, включительно справа
//
// [n:m] >< [n:m] — пересечение
//
// баунды всегда отсортированы по возрастанию нижней границы:
//
// [n:m][n:m] — конкретные значения не имеют значения, но m1 <= n2
// [0:3][1:2] — конкретные значения важны, баунд 1 входит в баунд 2

func Union[T any](s ...Span[T]) (res Span[T]) {
	if len(s) == 0 {
		return nil
	}

	for _, s := range s {
		if s == nil {
			continue
		} else if res == nil {
			res = s
		} else {
			res = res.Union(s)
		}
	}

	return res
}

func IsEqual[T comparable](a, b Span[T]) bool {
	if a == nil || b == nil {
		return a == b
	}

	aBounds, bBounds := a.Bounds(), b.Bounds()
	if len(aBounds) != len(bBounds) {
		return false
	}

	for i := range aBounds {
		if aBounds[i] != bBounds[i] {
			return false
		}
	}

	return true
}

type Span[T any] interface {
	// Search(T) Position
	Union(Span[T]) Span[T]
	UnionBound(Bound[T]) Span[T]

	Difference(Span[T]) Span[T]
	DifferenceBound(Bound[T]) Span[T]
	// // Contains checks, that all values of one span exists in other span
	// Contains(Span[T]) bool
	//
	// ContainsBound checks, that all values of one bound exists in other span
	ContainsBound(Bound[T]) bool
	//
	// Bounds returns a list of all bounds in a span
	Bounds() []Bound[T]
}

type span[T any] struct {
	// next is a function, which returns nearest next values
	next nextFunc[T]
	// cmp is a function, which compares two values. -1 means that a < b, 0
	// means that a == b, and +1 means that a > 0
	cmp compareFunc[T]
	// bounds is a list of bounds, ordered by by lower bound. this list
	// guarantees, that there is no any value, that contains in 2 bounds a the
	// same time
	bounds []Bound[T]
}

func NewInt(b ...Bound[int]) Span[int]             { return New(nextInt, cmp.Compare, b...) }
func NewInt8(b ...Bound[int8]) Span[int8]          { return New(nextInt, cmp.Compare, b...) }
func NewInt16(b ...Bound[int16]) Span[int16]       { return New(nextInt, cmp.Compare, b...) }
func NewInt32(b ...Bound[int32]) Span[int32]       { return New(nextInt, cmp.Compare, b...) }
func NewInt64(b ...Bound[int64]) Span[int64]       { return New(nextInt, cmp.Compare, b...) }
func NewUint(b ...Bound[uint]) Span[uint]          { return New(nextInt, cmp.Compare, b...) }
func NewUint8(b ...Bound[uint8]) Span[uint8]       { return New(nextInt, cmp.Compare, b...) }
func NewUint16(b ...Bound[uint16]) Span[uint16]    { return New(nextInt, cmp.Compare, b...) }
func NewUint32(b ...Bound[uint32]) Span[uint32]    { return New(nextInt, cmp.Compare, b...) }
func NewUint64(b ...Bound[uint64]) Span[uint64]    { return New(nextInt, cmp.Compare, b...) }
func NewFloat32(b ...Bound[float32]) Span[float32] { return New(math.Nextafter32, cmp.Compare, b...) }
func NewFloat64(b ...Bound[float64]) Span[float64] { return New(math.Nextafter, cmp.Compare, b...) }
func NewByte(b ...Bound[byte]) Span[byte]          { return New(nextInt, cmp.Compare, b...) }
func NewRune(b ...Bound[rune]) Span[rune]          { return New(nextInt, cmp.Compare, b...) }

// TODO: implement nextString in a correct way.
//
// Current problem is that we can get next value fomr nextString, but we can't
// do it for previous string value.
//
// Even though, it's worthless right now to spend so much time on this, because
// there are soooooo tiny amount of cases, when we need to use string as a span.
func _NewString(b ...Bound[string]) Span[string] { return New(nextString, cmp.Compare, b...) }

// just to tell staticcheck that we are using this function in the future
var _ = _NewString

type nextSimple interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~uintptr
}

func zero[T any]() (t T) { return t }
func nextInt[T nextSimple](v, t T) T {
	switch {
	case v == t:
		return v
	case v < t:
		return v + 1
	default:
		return v - 1
	}
}

func nextString(v, t string) string {
	switch {
	case v == t:
		return v
	case v < t:
		return string(nextArray([]rune(v), func(r rune) rune { return r + 1 }, zero[rune]))
	default:
		panic("not implemented")
	}
}

// using zero as function, cause for array of ints zero value will be negative minimum value
func nextArray[T cmp.Ordered](a []T, next func(T) T, zero func() T) []T {
	if len(a) == 0 {
		return []T{zero()}
	}

	for i := len(a) - 1; i >= 0; i-- {
		old := a[i]
		a[i] = next(old)
		if a[i] > old { // if not overflow
			break
		}

		// If it did overflow, carry over to the next element
		// If we're at the first element, insert a new element at the beginning
		if i == 0 {
			a = make([]T, len(a)+1)
			for i := range a {
				a[i] = zero()
			}
		}
	}

	return a
}

// func prevArray[T cmp.Ordered](a []T, prev func(T) T) []T {
// 	if len(a) == 0 {
// 		panic("got the smallest possible value")
// 	}
// 	for i := len(a) - 1; i >= 0; i-- {
// 		old := a[i]
// 		a[i] = prev(old)
// 		if a[i] < old { // if not underflow
// 			break
// 		}
// 		// If it did underflow, carry over to the next element
// 		// If we're at the first element, insert a new element at the beginning
// 		if i == 0 {
// 			a = append([]T{prev(old)}, a...)
// 		}
// 	}
// 	return a
// }

func New[T any](next nextFunc[T], cmp func(T, T) int, bounds ...Bound[T]) Span[T] {
	if cmp == nil {
		panic("cmp function is nil")
	} else if next == nil {
		panic("next function is nil")
	}

	var s Span[T] = span[T]{
		next:   next,
		cmp:    cmp,
		bounds: make([]Bound[T], 0, len(bounds)),
	}
	for _, b := range bounds {
		s = s.UnionBound(b)
	}

	return s
}

func ToBasic[T any](s Span[T]) [][2]Edge[T] {
	return slices.Remap(s.Bounds(), func(b Bound[T]) [2]Edge[T] { return [2]Edge[T]{b.Lo, b.Hi} })
}

func ToBasicBounds[T any](s ...Bound[T]) [][2]Edge[T] {
	return slices.Remap(s, func(b Bound[T]) [2]Edge[T] { return [2]Edge[T]{b.Lo, b.Hi} })
}

func FromBasicOrdered[T cmp.Ordered](s [][2]T) Span[T] {
	return New(nil, cmp.Compare, slices.Remap(s, func(b [2]T) Bound[T] { return NewBound(true, b[0], b[1], true) })...)
}

// MakeStrictBounds creates a new span with the given bounds, ensuring that all
// bounds have included edges. If some bound in input span contains excluded
// edge, `next` function will be used to get the next value for the bound.
func MakeStrictBounds[T any](s Span[T], cmp compareFunc[T], next nextFunc[T]) Span[T] {
	bounds := s.Bounds()
	if len(bounds) == 0 {
		return s
	}

	minValue := bounds[0].Lo.Value
	maxValue := bounds[len(bounds)-1].Hi.Value

	newBounds := make([]Bound[T], 0, len(bounds))
	for _, bound := range bounds {
		if bound.Lo.Included && bound.Hi.Included {
			newBounds = append(newBounds, bound)
			continue
		}

		nlo, nhi := bound.Lo.Value, bound.Hi.Value

		if !bound.Lo.Included {
			nlo = next(bound.Lo.Value, maxValue)
		}
		if !bound.Hi.Included {
			nhi = next(bound.Hi.Value, minValue)
		}

		// handling invalid bound
		if cmp(nlo, nhi) > 0 {
			continue
		}

		newBounds = append(newBounds, NewBoundEdgesFunc(Edge[T]{Value: nlo, Included: true}, Edge[T]{Value: nhi, Included: true}, cmp))
	}

	return New(next, cmp, newBounds...)
}

func (s span[T]) Bounds() []Bound[T] { return s.bounds }

func (s span[T]) ContainsBound(y Bound[T]) bool {
	if i, ok := s.search(y.Lo.Value); ok {
		return s.bounds[i].Contains(s.cmp, y)
	}

	return false
}

func (s span[T]) Union(y Span[T]) (z Span[T]) {
	z = s
	for _, b := range y.Bounds() {
		z = z.UnionBound(b)
	}

	return z
}

func (s span[T]) UnionBound(bound Bound[T]) Span[T] {
	var newBounds []Bound[T]

	// Iterate over existing bounds in the interval
	for _, existingBound := range s.bounds {
		// If they overlap, merge them and replace the existing bound with the merged one
		if mergedBound, merged := UnionBounds(s.next, s.cmp, existingBound, bound); merged {
			bound = mergedBound
			continue
		}
		// If there's no overlap, keep the existing bound unchanged
		newBounds = append(newBounds, existingBound)
	}

	// Add the new bound to the interval
	newBounds = append(newBounds, bound)

	// Sort the new bounds by lower bound
	sort.Slice(newBounds, func(i, j int) bool {
		return s.cmp(newBounds[i].Lo.Value, newBounds[j].Lo.Value) == -1
	})

	// Update the interval's bounds
	s.bounds = newBounds

	return s
}

func (s span[T]) Difference(y Span[T]) (z Span[T]) {
	z = s
	for _, b := range y.Bounds() {
		z = z.DifferenceBound(b)
	}

	return z
}

func (s span[T]) DifferenceBound(y Bound[T]) Span[T] {
	var newBounds []Bound[T]

	// Iterate over existing bounds in the interval
	for _, existingBound := range s.bounds {
		// If the existing bound is the one to be removed, skip it
		if y.Contains(s.cmp, existingBound) {
			continue
		}
		// If the existing bound overlaps with the bound to be removed, split it
		if existingBound.Overlaps(s.cmp, y) {
			// Get the difference between the existing bound and the bound to be removed
			diffBounds := existingBound.Difference(s.cmp, y)
			// Add the difference bounds to the new bounds list
			newBounds = append(newBounds, diffBounds...)
			continue
		}
		// If there's no overlap, keep the existing bound unchanged
		newBounds = append(newBounds, existingBound)
	}

	// Update the interval's bounds
	s.bounds = newBounds

	return s
}

func (s span[T]) search(t T) (int, bool) {
	return slices.BinarySearchFunc(s.bounds, t, func(a Bound[T], b T) int { return a.Position(s.cmp, b) })
}

// func (s span[T]) Search(value T) Position {
// 	if len(s.bounds) == 0 {
// 		return PositionNowhere{}
// 	} else if i, ok := s.search(value); ok {
// 		return PositionExact(i)
// 	} else if i == 0 {
// 		return PositionLower{}
// 	} else if i == len(s.bounds) {
// 		return PositionHigher{}
// 	} else {
// 		return PositionBetween{Lo: i - 1, Hi: i}
// 	}
// }

// type Position interface{ _Position() }
//
// type PositionNowhere struct{}
// type PositionHigher struct{}
// type PositionLower struct{}
// type PositionExact int
// type PositionBetween struct{ Lo, Hi int }
//
// func (PositionNowhere) _Position()       {}
// func (PositionNowhere) String() string   { return "PositionNowhere{}" }
// func (PositionHigher) _Position()        {}
// func (PositionHigher) String() string    { return "PositionHigher{}" }
// func (PositionLower) _Position()         {}
// func (PositionLower) String() string     { return "PositionLower{}" }
// func (PositionExact) _Position()         {}
// func (p PositionExact) String() string   { return fmt.Sprintf("PositionExact{%v}", int(p)) }
// func (PositionBetween) _Position()       {}
// func (p PositionBetween) String() string { return fmt.Sprintf("PositionBetween{%v : %v}", p.Lo, p.Hi) }

func (s span[T]) String() string { return joinStringer(s.bounds, "") }

func (s span[T]) Format(f fmt.State, verb rune) {
	flags := string(slices.Filter([]rune("-+# 0"), func(r rune) bool { return f.Flag(int(r)) }))
	fmtValue := "%" + flags + string(verb)

	for _, b := range s.bounds {
		fmt.Fprintf(f, fmtValue, b)
	}
}

func joinStringer[S ~[]T, T fmt.Stringer](s S, sep string) string {
	return strings.Join(slices.Remap(s, func(s T) string { return s.String() }), sep)
}
