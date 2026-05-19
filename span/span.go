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

	"github.com/quenbyako/ext/itertools"
	"github.com/quenbyako/ext/maps"
	"github.com/quenbyako/ext/set"
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
// [n1:m1][n2:m2] — конкретные значения не имеют значения, но m1 <= n2
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
	equal := true
	itertools.Zip(itertools.Values(aBounds), itertools.Values(bBounds))(func(a, b Bound[T]) bool {
		if a != b {
			equal = false
			return false
		}
		return true
	})
	return equal
}

type Span[T any] interface {
	// Search(T) Position
	Union(s Span[T]) Span[T]
	UnionBound(b Bound[T]) Span[T]

	Subtract(s Span[T]) Span[T]
	SubtractBound(b Bound[T]) Span[T]
	// Contains checks, that all values of one span exists in other span
	Contains(s Span[T]) bool

	// ContainsBound checks, that all values of one bound exists in other span
	ContainsBound(b Bound[T]) bool

	// Bounds returns a list of all bounds in a span. The bounds are ordered by
	// their lower bound.
	Bounds() iterSeq2[int, Bound[T]]

	Search(t T) (int, bool)
}

type span[T any] struct {
	// next is a function, which returns nearest next values
	next nextFunc[T]

	// cmp is a function, which compares two values. -1 means that a < b, 0
	// means that a == b, and +1 means that a > 0
	cmp cmpFunc[T]

	// bounds is a list of bounds, ordered by lower bound. this list
	// guarantees, that there is no any value, that contains in 2 bounds a the
	// same time
	bounds []Bound[T]
}

func NewInt(b ...Bound[int]) Span[int] {
	return New(NextInt[int], cmp.Compare[int], b...)
}

func NewInt8(b ...Bound[int8]) Span[int8] {
	return New(NextInt[int8], cmp.Compare[int8], b...)
}

func NewInt16(b ...Bound[int16]) Span[int16] {
	return New(NextInt[int16], cmp.Compare[int16], b...)
}

func NewInt32(b ...Bound[int32]) Span[int32] {
	return New(NextInt[int32], cmp.Compare[int32], b...)
}

func NewInt64(b ...Bound[int64]) Span[int64] {
	return New(NextInt[int64], cmp.Compare[int64], b...)
}

func NewUint(b ...Bound[uint]) Span[uint] {
	return New(NextInt[uint], cmp.Compare[uint], b...)
}

func NewUint8(b ...Bound[uint8]) Span[uint8] {
	return New(NextInt[uint8], cmp.Compare[uint8], b...)
}

func NewUint16(b ...Bound[uint16]) Span[uint16] {
	return New(NextInt[uint16], cmp.Compare[uint16], b...)
}

func NewUint32(b ...Bound[uint32]) Span[uint32] {
	return New(NextInt[uint32], cmp.Compare[uint32], b...)
}

func NewUint64(b ...Bound[uint64]) Span[uint64] {
	return New(NextInt[uint64], cmp.Compare[uint64], b...)
}

func NewFloat32(b ...Bound[float32]) Span[float32] {
	return New(math.Nextafter32, cmp.Compare[float32], b...)
}

func NewFloat64(b ...Bound[float64]) Span[float64] {
	return New(math.Nextafter, cmp.Compare[float64], b...)
}

func NewByte(b ...Bound[byte]) Span[byte] {
	return New(NextInt[byte], cmp.Compare[byte], b...)
}

func NewRune(b ...Bound[rune]) Span[rune] {
	return New(NextInt[rune], cmp.Compare[rune], b...)
}

// TODO: implement nextString in a correct way.
//
// Current problem is that we can get next value form nextString, but we can't
// do it for previous string value.
//
// Even though, it's worthless right now to spend so much time on this, because
// there are soooooo tiny amount of cases, when we need to use string as a span.
func _NewString(b ...Bound[string]) Span[string] { return New(nextString, cmp.Compare[string], b...) }

// just to tell staticcheck that we are using this function in the future.
var _ = _NewString

type nextSimple interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~uintptr
}

func zero[T any]() (t T) { return t }
func NextInt[T nextSimple](v, t T) T {
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

// using zero as function, cause for array of ints zero value will be negative minimum value.
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

func ToBasic[T any](s Span[T]) iterSeq2[int, [2]Edge[T]] {
	return itertools.RemapPairs(s.Bounds(), func(i int, b Bound[T]) (int, [2]Edge[T]) { return i, [2]Edge[T]{b.Lo, b.Hi} })
}

func ToBasicBounds[T any](s ...Bound[T]) [][2]Edge[T] {
	return slices.Remap(s, func(b Bound[T]) [2]Edge[T] { return [2]Edge[T]{b.Lo, b.Hi} })
}

func FromBasicOrdered[T cmp.Ordered](s [][2]T) Span[T] {
	return New(nil, compare[T], slices.Remap(s, func(b [2]T) Bound[T] { return NewBound(true, b[0], b[1], true) })...)
}

// MakeStrictBounds creates a new span with the given bounds, ensuring that all
// bounds have included edges. If some bound in input span contains excluded
// edge, `next` function will be used to get the next value for the bound.
func MakeStrictBounds[T any](s Span[T], cmp cmpFunc[T], next nextFunc[T]) Span[T] {
	newBounds := make([]Bound[T], 0)

	s.Bounds()(func(_ int, bound Bound[T]) bool {
		if bound.Lo.Included && bound.Hi.Included {
			newBounds = append(newBounds, bound)

			return true
		}

		nlo, nhi := bound.Lo.Value, bound.Hi.Value

		if !bound.Lo.Included {
			nlo = next(bound.Lo.Value, bound.Hi.Value)
		}

		if !bound.Hi.Included {
			nhi = next(bound.Hi.Value, bound.Lo.Value)
		}

		// handling invalid bound
		if cmp(nlo, nhi) > 0 {
			return true
		}

		newBounds = append(newBounds, NewBoundEdgesFunc(
			Edge[T]{Value: nlo, Included: true},
			Edge[T]{Value: nhi, Included: true},
			cmp,
		))
		return true
	})

	return New(next, cmp, newBounds...)
}

func EachItem[T any](s Span[T], next nextFunc[T], cmp cmpFunc[T]) iterSeq[T] {
	bounds := s.Bounds()

	return func(yield func(T) bool) {
		bounds(func(_ int, b Bound[T]) bool {
			r := b.Lo.Value
			if !b.Lo.Included {
				r = next(b.Lo.Value, b.Hi.Value)
			}

			if b.Hi.Included {
				for {
					if !yield(r) {
						return false
					}

					if cmp(r, b.Hi.Value) >= 0 {
						break
					}

					r = next(r, b.Hi.Value)
				}
			} else {
				for {
					if !yield(r) {
						return false
					}

					r = next(r, b.Hi.Value)

					if cmp(r, b.Hi.Value) >= 0 {
						break
					}
				}
			}
			return true
		})
	}
}

func (s span[T]) Bounds() iterSeq2[int, Bound[T]] { return slices.All(s.bounds) }

func (s span[T]) Contains(y Span[T]) bool {
	contains := true
	y.Bounds()(func(_ int, b Bound[T]) bool {
		if !s.ContainsBound(b) {
			contains = false
			return false
		}
		return true
	})

	return contains
}

func (s span[T]) ContainsBound(y Bound[T]) bool {
	if i, ok := s.Search(y.Lo.Value); ok {
		return s.bounds[i].Contains(s.cmp, y)
	}

	return false
}

func (s span[T]) Union(y Span[T]) (z Span[T]) {
	z = s
	y.Bounds()(func(_ int, b Bound[T]) bool {
		z = z.UnionBound(b)
		return true
	})

	return z
}

func (s span[T]) UnionBound(bound Bound[T]) Span[T] {
	newBounds := make([]Bound[T], 0)

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

func (s span[T]) Subtract(y Span[T]) (z Span[T]) {
	z = s
	y.Bounds()(func(_ int, b Bound[T]) bool {
		z = z.SubtractBound(b)
		return true
	})

	return z
}

func (s span[T]) SubtractBound(y Bound[T]) Span[T] {
	newBounds := make([]Bound[T], 0, len(s.bounds))

	// Iterate over existing bounds in the interval
	for _, existingBound := range s.bounds {
		// If the existing bound is the one to be removed, skip it
		if y.Contains(s.cmp, existingBound) {
			continue
		}
		// If the existing bound overlaps with the bound to be removed, split it
		if existingBound.Overlaps(s.cmp, y) {
			// Get the difference between the existing bound and the bound to be removed
			diffBounds := existingBound.Subtract(s.cmp, y)
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

func (s span[T]) Search(t T) (int, bool) {
	return slices.BinarySearchFunc(s.bounds, t, func(a Bound[T], b T) int { return a.Position(s.cmp, b) })
}

func (s span[T]) String() string { return fmt.Sprintf("%v", s) }

func (s span[T]) Format(f fmt.State, verb rune) {
	fmtValue := fmt.FormatString(f, verb)

	for _, b := range s.bounds {
		fmt.Fprintf(f, fmtValue, b)
	}
}


// ........................   0    1    2
// Example of sortedEdge [-Inf:0)[0:0](0:+Inf]
// will be
//
//	sortedEdges{
//	  excludedEnd:   [0]
//	  includedStart: [1]
//	  includedEnd:   [1]
//	  excludedStart: [2]
type sortedEdges struct {
	excludedEnd   []int
	includedStart []int
	includedEnd   []int
	excludedStart []int
}

// Split breaks set of ranges into a set of non-intersecting bounds, so
// that each range in the set is a sum of some of the bounds. returns not just
// non-intersecting bounds, but also indexes of bounds that are containing this
// range.
//
// examples:
//
//	0: Split([1:3][6:7], [2:3][5:6]) {
//	      return {
//	          [1:2) -> [0]
//	          [2:3] -> [0, 1]
//	          [5:6) -> [1]
//	          [6:6] -> [0, 1]
//	          (6:7] -> [0]
//	      }
//	   }
func Split[T comparable](spans []Span[T], cmp cmpFunc[T], next nextFunc[T]) (res map[Bound[T]][]int) {
	// 1. Собираем все уникальные границы из всех диапазонов.
	edges := collectEdges(spans)

	if len(edges) == 0 {
		return map[Bound[T]][]int{}
	}

	edgeKeys := maps.Keys(edges)
	slices.SortFunc(edgeKeys, cmp)

	res = make(map[Bound[T]][]int)

	var lastOpenedEdge Edge[T]
	currentSpans := set.New[int]()

	appendBound := func(edge Edge[T]) {
		// Создаём фрагмент
		fragmentBound := Bound[T]{
			Lo: lastOpenedEdge,
			Hi: edge,
		}

		// checking that bound has at least one value in it
		if !lastOpenedEdge.Included && fragmentBound.Position(cmp, next(lastOpenedEdge.Value, edge.Value)) != 0 {
			return
		}

		if currentSpans.Len() == 0 {
			panic("impossible case: no spans")
		}

		// Добавляем фрагмент в результат
		values := set.AsList(currentSpans)
		slices.Sort(values)

		res[fragmentBound] = values
	}

	for _, value := range edgeKeys {
		this := edges[value]

		hasExcludedEnd := len(this.excludedEnd) > 0
		hasIncludedStart := len(this.includedStart) > 0
		hasIncludedEnd := len(this.includedEnd) > 0
		hasExcludedStart := len(this.excludedStart) > 0

		atLeastOneElem := (hasExcludedEnd || hasIncludedStart || hasIncludedEnd || hasExcludedStart)
		if !atLeastOneElem {
			panic("impossible: bound has no elements")
		}

		// starting of bound
		if currentSpans.Len() == 0 {
			if hasExcludedEnd {
				panic("impossible: closing bound that was not opened")
			}

			if !hasIncludedStart {
				if hasIncludedEnd {
					panic("impossible: closing bound that was not opened")
				}

				lastOpenedEdge = Edge[T]{Value: value, Included: false}
				currentSpans = currentSpans.Add(this.excludedStart...)

				continue
			}

			lastOpenedEdge = Edge[T]{Value: value, Included: true}
			currentSpans = currentSpans.Add(this.includedStart...)

			if hasExcludedStart || hasIncludedEnd {
				appendBound(Edge[T]{Value: value, Included: true})
				currentSpans = currentSpans.Del(this.includedEnd...)

				lastOpenedEdge = Edge[T]{Value: value, Included: false}
			}

			currentSpans = currentSpans.Add(this.excludedStart...)

			continue
		}

		// handling previously opened bound
		if hasExcludedEnd || !hasExcludedEnd && hasIncludedStart {
			appendBound(Edge[T]{Value: value, Included: false})
		}

		currentSpans = currentSpans.Del(this.excludedEnd...)

		if hasExcludedEnd {
			lastOpenedEdge = Edge[T]{Value: value, Included: true}

			if !hasIncludedStart {
				if hasIncludedEnd {
					appendBound(Edge[T]{Value: value, Included: true})
				} else if !hasExcludedStart {
					continue
				} else if currentSpans.Len() > 0 {
					appendBound(Edge[T]{Value: value, Included: true})
				}
			}

			currentSpans = currentSpans.Add(this.includedStart...)

			if hasIncludedStart && ((!hasIncludedEnd && hasExcludedStart) || hasIncludedEnd) {
				appendBound(Edge[T]{Value: value, Included: true})
			}
		} else {
			if hasIncludedStart {
				lastOpenedEdge = Edge[T]{Value: value, Included: true}
			}

			currentSpans = currentSpans.Add(this.includedStart...)

			if !hasIncludedStart || (!hasIncludedEnd && hasExcludedStart) || hasIncludedEnd {
				appendBound(Edge[T]{Value: value, Included: true})
			}
		}

		currentSpans = currentSpans.Del(this.includedEnd...)

		if !hasIncludedStart || hasExcludedStart || hasIncludedEnd {
			lastOpenedEdge = Edge[T]{Value: value, Included: false}
		}

		currentSpans = currentSpans.Add(this.excludedStart...)
	}

	if currentSpans.Len() > 0 {
		panic("impossible case: some bounds are not closed")
	}

	return res
}

func collectEdges[T comparable](spans []Span[T]) map[T]sortedEdges {
	edges := make(map[T]sortedEdges, 0)
	for si, s := range spans {
		if s == nil {
			continue
		}

		s.Bounds()(func(_ int, b Bound[T]) bool {
			if _, ok := edges[b.Lo.Value]; !ok {
				edges[b.Lo.Value] = sortedEdges{}
			}

			lo := edges[b.Lo.Value]

			if b.Lo.Included {
				lo.includedStart = append(lo.includedStart, si)
			} else {
				lo.excludedStart = append(lo.excludedStart, si)
			}

			edges[b.Lo.Value] = lo

			if _, ok := edges[b.Hi.Value]; !ok {
				edges[b.Hi.Value] = sortedEdges{}
			}

			hi := edges[b.Hi.Value]

			if b.Hi.Included {
				hi.includedEnd = append(hi.includedEnd, si)
			} else {
				hi.excludedEnd = append(hi.excludedEnd, si)
			}

			edges[b.Hi.Value] = hi
			return true
		})
	}

	return edges
}
