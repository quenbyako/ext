// Copyright (c) 2020-2024 Richard Cooper
//
// This file is a part of quenbyako/ext package.
// See https://github.com/quenbyako/ext/blob/master/LICENSE for details

package span

import (
	"cmp"
	"errors"
	"fmt"
	"strings"
)

type nextFunc[T any] func(T) T
type compareFunc[T any] func(T, T) int

func IsBoundEqual[T comparable](a, b Bound[T]) bool {
	if a == nil || b == nil {
		return a == b
	}

	return a.Lo() == b.Lo() && a.Hi() == b.Hi()
}

// returns new bound, if bounds overlaps and it's possible to merge.
// Otherwise, returns copy of current bound and false.
func UnionBounds[T any](next func(T) T, cmp func(T, T) int, a, b Bound[T]) (Bound[T], bool) {
	if a.Contains(cmp, b) {
		return NewBoundEdgesFunc(a.Lo(), a.Hi(), cmp), true
	} else if b.Contains(cmp, a) {
		return NewBoundEdgesFunc(b.Lo(), b.Hi(), cmp), true
	}

	alo, ahi := a.Lo(), a.Hi()
	blo, bhi := b.Lo(), b.Hi()

	if !a.Overlaps(cmp, b) {
		// There are only 2 cases when the bounds can be connected, but they do
		// not intersect:
		//
		// 1) if the bounds are limited at one point, but only one of them does
		//    not include this point (example: [1:2] >< (2:3] -> [1:3])
		// 2) if both bounds include boundary points, and the upper bound of the
		//    lower bound + 1 is equal to the lower bound of the upper one
		//    (example: [1:2] >< [3:4] -> [1:4]. This is example only for int,
		//    for float the smallest possible value is added)
		switch {
		case cmp(ahi.Value, blo.Value) <= 0 && IsEdgeNear(next, cmp, ahi, blo): // left join
			return NewBoundEdgesFunc(alo, bhi, cmp), true
		case cmp(bhi.Value, alo.Value) <= 0 && IsEdgeNear(next, cmp, bhi, alo): // right join
			return NewBoundEdgesFunc(blo, ahi, cmp), true
		default: // all other  cases
			return NewBoundEdgesFunc(a.Lo(), a.Hi(), cmp), false
		}
	}

	lo, hi := minEdge(alo, blo, cmp), maxEdge(ahi, bhi, cmp)

	// now it overlaps by part, so bound must be modified
	return NewBoundEdgesFunc(lo, hi, cmp), true
}

type Bound[T any] interface {
	Lo() Edge[T]
	Hi() Edge[T]

	// Contains checks, that all values of b bound exists in a bound
	Contains(compareFunc[T], Bound[T]) bool
	Overlaps(compareFunc[T], Bound[T]) bool

	// Position compares single value with bound, and returns its position.
	//
	// Returns 0, if bound contains this value, +1, if `i` is less than higher
	// bound edge (according to `cmp` package, comparing `bound > i` should
	// return +1), and -1, if `i` is larger than lower bound edge (same idea:
	// for `bound < i` should return -1)
	Position(compareFunc[T], T) int

	Difference(compareFunc[T], Bound[T]) []Bound[T]

	fmt.Stringer
}

// принцип имплементации всех баундов:
//
// так как мы хотим сделать баунды для любых типов (от time.Time, до нод в
// kubernetes), нам нужно иметь функцию cmp, сравнивающую значения.
//
// ключевое вычисление в действиях над баундами — понимание расположение верхних
// и нижних границ каждого баунда, поэтому почти во всех методах идентичный
// шаблон:
//
//	b := bound[T]{lo: b.Lo(), hi: b.Hi()}
//	lolocmp := a.cmp(a.lo.Value, b.lo.Value)
//	lohicmp := x.cmp(a.lo.Value, b.hi.Value)
//	hihicmp := a.cmp(a.hi.Value, b.hi.Value)
//	hilocmp := x.cmp(a.hi.Value, b.lo.Value)
//
// каждое из сравнений используется только по требованию.
//
// `b` в этом шаблоне используется просто для красоты, т.к. не занимает лишнего
// места (добавляется лишь поинтер), а выглядит гораздо понятнее
type bound[T any] struct{ lo, hi Edge[T] }

var _ Bound[int] = bound[int]{}

var ErrInvalidBound = errors.New("bound value is invalid")

func ParseBound[T any](s string, parser func(s string) (T, error)) (_ Bound[T], err error) {
	b := bound[T]{}

	if len(s) < 5 {
		return nil, ErrInvalidBound
	}

	switch s[0] {
	case '[':
		b.lo.Included = true
	case '(':
		b.lo.Included = false
	default:
		return nil, ErrInvalidBound
	}
	switch s[len(s)-1] {
	case ']':
		b.hi.Included = true
	case ')':
		b.hi.Included = false
	default:
		return nil, ErrInvalidBound
	}

	if divider := strings.IndexRune(s, ':'); divider < 0 {
		return nil, ErrInvalidBound
	} else if b.lo.Value, err = parser(s[1:divider]); err != nil {
		return nil, err
	} else if b.hi.Value, err = parser(s[divider+1 : len(s)-1]); err != nil {
		return nil, err
	}

	return b, nil
}

func NewBoundXX[T cmp.Ordered](lo, hi T) Bound[T] { return NewBound(false, lo, hi, false) }
func NewBoundXI[T cmp.Ordered](lo, hi T) Bound[T] { return NewBound(false, lo, hi, true) }
func NewBoundIX[T cmp.Ordered](lo, hi T) Bound[T] { return NewBound(true, lo, hi, false) }
func NewBoundII[T cmp.Ordered](lo, hi T) Bound[T] { return NewBound(true, lo, hi, true) }
func NewBound[T cmp.Ordered](loIncluded bool, lo, hi T, hiIncluded bool) Bound[T] {
	return NewBoundEdges(newEdge(lo, loIncluded), newEdge(hi, hiIncluded))
}

func NewBoundEdges[T cmp.Ordered](lo, hi Edge[T]) Bound[T] {
	return NewBoundEdgesFunc(lo, hi, cmp.Compare)
}

// The implementation guarantees, that `cmp` function will be called AT MOST 4
// times for each method call.
func NewBoundEdgesFunc[T any](lo, hi Edge[T], cmp compareFunc[T]) Bound[T] {
	if cmp == nil {
		panic("cmp function is nil")
	} else if compared := cmp(lo.Value, hi.Value); compared > 0 {
		panic(fmt.Sprintf("lo is higher than hi: %v > %v", lo.Value, hi.Value))
	} else if compared == 0 && (!lo.Included || !hi.Included) {
		panic("no values inside bound: " + boundString(lo, hi))
	}

	return bound[T]{lo: lo, hi: hi}
}

func (x bound[T]) Lo() Edge[T] { return x.lo }
func (x bound[T]) Hi() Edge[T] { return x.hi }
func (a bound[T]) Contains(cmp compareFunc[T], b Bound[T]) bool {
	blo, bhi := b.Lo(), b.Hi()
	locmp := cmp(a.lo.Value, blo.Value)
	hicmp := cmp(a.hi.Value, bhi.Value)

	// Check if the start of b is within a
	startWithinA := locmp < 0 || (locmp == 0 && (a.lo.Included || blo.Included == a.lo.Included))
	// Check if the end of b is within a
	endWithinA := hicmp > 0 || (hicmp == 0 && (a.hi.Included || bhi.Included == a.hi.Included))

	return startWithinA && endWithinA
}

func (a bound[T]) Overlaps(cmp compareFunc[T], b Bound[T]) bool {
	blo, bhi := b.Lo(), b.Hi()
	lohicmp := cmp(a.lo.Value, bhi.Value)
	hilocmp := cmp(a.hi.Value, blo.Value)

	bTooLow := lohicmp > 0 || (lohicmp == 0 && (!a.lo.Included || a.lo.Included != bhi.Included))
	bTooHigh := hilocmp < 0 || (hilocmp == 0 && (!a.hi.Included || a.hi.Included != blo.Included))

	return !bTooHigh && !bTooLow

}

func (x bound[T]) Position(cmp compareFunc[T], i T) int {
	locmp := cmp(x.lo.Value, i)
	hicmp := cmp(x.hi.Value, i)

	if locmp > 0 || (locmp == 0 && !x.lo.Included) { // before low
		return +1
	} else if hicmp < 0 || (hicmp == 0 && !x.hi.Included) { // after high
		return -1
	} else {
		return 0
	}
}

func (a bound[T]) Difference(cmp compareFunc[T], b Bound[T]) (res []Bound[T]) {
	blo, bhi := b.Lo(), b.Hi()

	if b.Contains(cmp, a) {
		return []Bound[T]{}
	} else if !a.Overlaps(cmp, b) {
		return []Bound[T]{a}
	}

	if locmp := cmp(blo.Value, a.lo.Value); locmp == 0 {
		if a.lo.Included && !blo.Included { // [1:3] - (1:2] = [1:1](2:3]
			edge := newEdge(a.lo.Value, true)
			res = append(res, NewBoundEdgesFunc(edge, edge, cmp))
		}
	} else if locmp > 0 { // [1:3] - [2:3) = [1:2)[3:3]
		res = append(res, NewBoundEdgesFunc(a.lo, newEdge(blo.Value, !blo.Included), cmp))
	}

	if hicmp := cmp(bhi.Value, a.hi.Value); hicmp == 0 {
		if a.hi.Included && !bhi.Included { // [1:3] - (1:2] = [1:1](2:3]
			edge := newEdge(a.hi.Value, true)
			res = append(res, NewBoundEdgesFunc(edge, edge, cmp))
		}
	} else if hicmp < 0 { // [1:3] - (1:2] = [1:1](2:3]
		res = append(res, NewBoundEdgesFunc(newEdge(bhi.Value, !bhi.Included), a.hi, cmp))
	}

	return res
}

func (x bound[T]) String() (res string) {
	return boundString(newEdge(x.lo.Value, x.lo.Included), newEdge(x.hi.Value, x.hi.Included))
}

func boundString[T any](lo, hi Edge[T]) (res string) {
	if lo.Included {
		res += "["
	} else {
		res += "("
	}

	res += fmt.Sprintf("%v:%v", lo.Value, hi.Value)

	if hi.Included {
		res += "]"
	} else {
		res += ")"
	}

	return res
}
