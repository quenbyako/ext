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

	"github.com/quenbyako/ext/slices"
)

type nextFunc[T any] func(T, T) T
type compareFunc[T any] func(T, T) int

// returns new bound, if bounds overlaps and it's possible to merge.
// Otherwise, returns copy of current bound and false.
func UnionBounds[T any](next nextFunc[T], cmp compareFunc[T], a, b Bound[T]) (Bound[T], bool) {
	if a.Contains(cmp, b) {
		return NewBoundEdgesFunc(a.Lo, a.Hi, cmp), true
	} else if b.Contains(cmp, a) {
		return NewBoundEdgesFunc(b.Lo, b.Hi, cmp), true
	}

	alo, ahi := a.Lo, a.Hi
	blo, bhi := b.Lo, b.Hi

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
			return a, false
		}
	}

	lo, hi := minEdge(alo, blo, cmp), maxEdge(ahi, bhi, cmp)

	// now it overlaps by part, so bound must be modified
	return NewBoundEdgesFunc(lo, hi, cmp), true
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
type Bound[T any] struct{ Lo, Hi Edge[T] }

var ErrInvalidBound = errors.New("bound value is invalid")

func ParseBound[T any](s string, parser func(s string) (T, error)) (_ Bound[T], err error) {
	b := Bound[T]{}

	if len(s) < 5 {
		return Bound[T]{}, ErrInvalidBound
	}

	switch s[0] {
	case '[':
		b.Lo.Included = true
	case '(':
		b.Lo.Included = false
	default:
		return Bound[T]{}, ErrInvalidBound
	}
	switch s[len(s)-1] {
	case ']':
		b.Hi.Included = true
	case ')':
		b.Hi.Included = false
	default:
		return Bound[T]{}, ErrInvalidBound
	}

	if divider := strings.IndexRune(s, ':'); divider < 0 {
		return Bound[T]{}, ErrInvalidBound
	} else if b.Lo.Value, err = parser(s[1:divider]); err != nil {
		return Bound[T]{}, err
	} else if b.Hi.Value, err = parser(s[divider+1 : len(s)-1]); err != nil {
		return Bound[T]{}, err
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
		panic("no values inside bound: " + boundString(lo, hi, "", 'v'))
	}

	return Bound[T]{Lo: lo, Hi: hi}
}

func (a Bound[T]) Contains(cmp compareFunc[T], b Bound[T]) bool {

	locmp := cmp(a.Lo.Value, b.Lo.Value)
	hicmp := cmp(a.Hi.Value, b.Hi.Value)

	// Check if the start of b is within a
	startWithinA := locmp < 0 || (locmp == 0 && (a.Lo.Included || b.Lo.Included == a.Lo.Included))
	// Check if the end of b is within a
	endWithinA := hicmp > 0 || (hicmp == 0 && (a.Hi.Included || b.Hi.Included == a.Hi.Included))

	return startWithinA && endWithinA
}

func (a Bound[T]) Overlaps(cmp compareFunc[T], b Bound[T]) bool {
	lohicmp := cmp(a.Lo.Value, b.Hi.Value)
	hilocmp := cmp(a.Hi.Value, b.Lo.Value)

	bTooLow := lohicmp > 0 || (lohicmp == 0 && (!a.Lo.Included || a.Lo.Included != b.Hi.Included))
	bTooHigh := hilocmp < 0 || (hilocmp == 0 && (!a.Hi.Included || a.Hi.Included != b.Lo.Included))

	return !bTooHigh && !bTooLow

}

func (x Bound[T]) Position(cmp compareFunc[T], i T) int {
	locmp := cmp(x.Lo.Value, i)
	hicmp := cmp(x.Hi.Value, i)

	if locmp > 0 || (locmp == 0 && !x.Lo.Included) { // before low
		return +1
	} else if hicmp < 0 || (hicmp == 0 && !x.Hi.Included) { // after high
		return -1
	} else {
		return 0
	}
}

func (a Bound[T]) Difference(cmp compareFunc[T], b Bound[T]) (res []Bound[T]) {
	blo, bhi := b.Lo, b.Hi

	if b.Contains(cmp, a) {
		return []Bound[T]{}
	} else if !a.Overlaps(cmp, b) {
		return []Bound[T]{a}
	}

	if locmp := cmp(blo.Value, a.Lo.Value); locmp == 0 {
		if a.Lo.Included && !blo.Included { // [1:3] - (1:2] = [1:1](2:3]
			edge := newEdge(a.Lo.Value, true)
			res = append(res, NewBoundEdgesFunc(edge, edge, cmp))
		}
	} else if locmp > 0 { // [1:3] - [2:3) = [1:2)[3:3]
		res = append(res, NewBoundEdgesFunc(a.Lo, newEdge(blo.Value, !blo.Included), cmp))
	}

	if hicmp := cmp(bhi.Value, a.Hi.Value); hicmp == 0 {
		if a.Hi.Included && !bhi.Included { // [1:3] - (1:2] = [1:1](2:3]
			edge := newEdge(a.Hi.Value, true)
			res = append(res, NewBoundEdgesFunc(edge, edge, cmp))
		}
	} else if hicmp < 0 { // [1:3] - (1:2] = [1:1](2:3]
		res = append(res, NewBoundEdgesFunc(newEdge(bhi.Value, !bhi.Included), a.Hi, cmp))
	}

	return res
}

func (x Bound[T]) String() (res string) {
	return boundString(newEdge(x.Lo.Value, x.Lo.Included), newEdge(x.Hi.Value, x.Hi.Included), "", 'v')
}

func (s Bound[T]) Format(f fmt.State, verb rune) {
	flags := string(slices.Filter([]rune("-+# 0"), func(r rune) bool { return f.Flag(int(r)) }))

	f.Write([]byte(boundString(s.Lo, s.Hi, flags, verb)))
}

func boundString[T any](lo, hi Edge[T], flags string, verb rune) (res string) {
	fmtValue := "%" + flags + string(verb)

	if lo.Included {
		res += "["
	} else {
		res += "("
	}

	res += fmt.Sprintf(fmtValue+":"+fmtValue, lo.Value, hi.Value)

	if hi.Included {
		res += "]"
	} else {
		res += ")"
	}

	return res
}
