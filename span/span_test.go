// Copyright (c) 2020-2024 Richard Cooper
//
// This file is a part of quenbyako/ext package.
// See https://github.com/quenbyako/ext/blob/master/LICENSE for details

package span_test

import (
	"cmp"
	"testing"
	"unicode"

	"github.com/quenbyako/ext/span"
	. "github.com/quenbyako/ext/span"
)

func s(b ...Bound[rune]) Span[rune]      { return New(Next[rune], cmp.Compare, b...) }
func b[T cmp.Ordered](lo, hi T) Bound[T] { return NewBoundII(lo, hi) }
func r[T cmp.Ordered](r T) Bound[T]      { return NewBoundII(r, r) }

func TestSearch(t *testing.T) {
	t.Skip()
	for _, tt := range []struct {
		// a    Span[rune]
		// r    rune
		// want Position
	}{
		// {Sr(b('a', 'a')), 'a', PositionExact(0)},
		// {Sr(b('a', 'z')), 'o', PositionExact(0)},
		// {Sr(b('a', 'z')), '|', PositionHigher{}},
		// {Sr(b('0', '9'), b('a', 'z')), '1', PositionExact(0)},
		// {Sr(b('0', '9'), b('a', 'z')), 'a', PositionExact(1)},
		// {Sr(b('0', '9'), b('a', 'z')), 'b', PositionExact(1)},
		// {Sr(b('0', '9'), b('a', 'z')), '!', PositionLower{}},
		// {Sr(b('0', '9'), b('a', 'z')), '@', PositionBetween{Lo: 0, Hi: 1}},
	} {
		t.Run("", func(t *testing.T) {
			_ = tt
			//	requireEqual(t, tt.want, tt.a.Search(tt.r))
		})
	}
}

func TestUnionSpans(t *testing.T) {
	for _, tt := range []struct{ a, b, want Span[rune] }{
		{s(), s(), s()},
		{s(), s(b('0', '9')), s(b('0', '9'))},
		{s(b('a', 'z')), s(), s(b('a', 'z'))},
		{s(b('a', 'z')), s(b('a', 'z')), s(b('a', 'z'))},
		{s(b('0', '9')), s(b('0', '9')), s(b('0', '9'))},
		{s(b('a', 'o')), s(b('o', 'z')), s(b('a', 'z'))},
		{s(b('a', 'p')), s(b('n', 'z')), s(b('a', 'z'))},
		{s(b('a', 't')), s(b('o', 'z')), s(b('a', 'z'))},
		{s(b('a', 't')), s(b('t', 'z')), s(b('a', 'z'))},
		{s(b('a', 'y')), s(b('b', 'z')), s(b('a', 'z'))},
		{s(b('a', 'z')), s(b('b', 'y')), s(b('a', 'z'))},
		{s(b('a', 'z')), s(b('b', 'y')), s(b('a', 'z'))},
		{s(b('a', 'z')), s(b('n', 'p')), s(b('a', 'z'))},
		{s(b('b', 'y')), s(b('a', 'z')), s(b('a', 'z'))},
		{s(b('b', 'y')), s(b('a', 'z')), s(b('a', 'z'))},
		{s(b('b', 'z')), s(b('a', 'y')), s(b('a', 'z'))},
		{s(b('n', 'p')), s(b('a', 'z')), s(b('a', 'z'))},
		{s(b('o', 'z')), s(b('a', 'o')), s(b('a', 'z'))},
		{s(b('o', 'z')), s(b('a', 't')), s(b('a', 'z'))},
		{s(b('t', 'z')), s(b('a', 't')), s(b('a', 'z'))},
		{s(b('a', 'a')), s(r('c')), s(b('a', 'a'), b('c', 'c'))},
		{s(b('a', 'z')), s(r('A')), s(b('A', 'A'), b('a', 'z'))},
		{s(b('c', 'z')), s(r('a')), s(b('a', 'a'), b('c', 'z'))},
		{s(b('0', '9')), s(b('a', 'z')), s(b('0', '9'), b('a', 'z'))},
		{s(b('0', '9')), s(b('a', 'z')), s(b('0', '9'), b('a', 'z'))},
		{s(b('a', 'd')), s(b('d', 'f'), b('f', 'i')), s(b('a', 'i'))},
		{s(b('a', 'n'), b('p', 'z')), s(b('n', 'p')), s(b('a', 'z'))},
		{s(b('a', 't')), s(b('x', 'z')), s(b('a', 't'), b('x', 'z'))},
		{s(b('a', 'z')), s(b('0', '9')), s(b('0', '9'), b('a', 'z'))},
		{s(b('a', 'z')), s(b('0', '9')), s(b('0', '9'), b('a', 'z'))},
		{s(b('a', 'c')), s(b('d', 'f'), b('g', 'i')), s(b('a', 'c'), b('d', 'f'), b('g', 'i'))},
		{s(b('A', 'J'), b('a', 'j'), b('l', 'r')), s(r('L')), s(b('A', 'J'), b('L', 'L'), b('a', 'j'), b('l', 'r'))},
	} {
		t.Run("", compareSpan(tt.want, tt.a.Union(tt.b)))
	}
}

func TestDifferenceSpans(t *testing.T) {
	for _, tt := range []struct{ a, b, want Span[int] }{
		{Si(bli("[1:6]")...), Si(bli("[2:4]")...), Si(bli("[1:2) (4:6]")...)},
		{Si(bli("[1:6]")...), Si(bli("[2:2]")...), Si(bli("[1:2) (2:6]")...)},
		{Si(bli("[1:3) [4:6]")...), Si(bli("[3:4]")...), Si(bli("[1:3) (4:6]")...)},
	} {
		t.Run("", compareSpan(tt.want, tt.a.Difference(tt.b)))
	}
}

type TestRunner interface {
	Name() string
	Run(t *testing.T)
}

type TestUnionNearCase[T comparable] struct {
	a    Span[T]
	b    Bound[T]
	want Span[T]
}

func (TestUnionNearCase[T]) Name() string { return "" }

func (tt TestUnionNearCase[T]) Run(t *testing.T) {
	requireEqualSpan(t, tt.want, tt.a.UnionBound(tt.b))
}

func TestUnionNear(t *testing.T) {
	for _, tt := range []TestRunner{
		TestUnionNearCase[int]{a: Si(b(1, 6)), b: b(3, 4), want: Si(b(1, 6))},
		TestUnionNearCase[int]{a: Si(b(3, 4)), b: b(1, 6), want: Si(b(1, 6))},
		TestUnionNearCase[int]{a: Si(b(1, 2), b(3, 4)), b: b(2, 3), want: Si(b(1, 4))},
		TestUnionNearCase[int]{a: Si(b(1, 2), b(5, 6)), b: b(3, 4), want: Si(b(1, 6))},
	} {
		t.Run(tt.Name(), tt.Run)
	}
}

func TestFold(t *testing.T) {
	for _, tt := range []struct {
		in   Span[rune]
		want Span[rune]
	}{
		{s(b('0', '9')), s(b('0', '9'))},
		{s(b('a', 'j')), s(b('A', 'J'), b('a', 'j'))},
		{s(b('a', 'j'), b('l', 'r')), s(b('A', 'J'), b('L', 'R'), b('a', 'j'), b('l', 'r'))},
		{s(b('a', 'j'), b('l', 'r'), b('t', 'z')), s(b('A', 'J'), b('L', 'R'), b('T', 'Z'), b('a', 'j'), b('l', 'r'), b('t', 'z'))},
		{s(b('0', '9'), b('a', 'z')), s(b('0', '9'), b('A', 'Z'), b('a', 'z'))},
	} {
		t.Run("", compareSpan(tt.want, fold(tt.in)))
	}
}

func TestMakeStrictBounds(t *testing.T) {
	for _, tt := range []struct {
		in   Span[rune]
		want Span[rune]
	}{
		{sr("(a:z)"), sr("[b:y]")},
		{sr("[a:z)"), sr("[a:y]")},
		{sr("[a:z)", "[0:9]"), sr("[a:y]", "[0:9]")},
		{sr("(a:z)", "(0:9)"), sr("[b:y]", "[1:8]")},

		// special cases:
		// * MakeStrictBounds doesn't normalizing,
		{sr("[1:2]", "[2:3]"), sr("[1:2]", "[2:3]")},
		// * cuts invalid bounds,
		{NewRune(Bound[rune]{Edge[rune]{Value: 1, Included: false}, Edge[rune]{Value: 2, Included: false}}), span.NewRune()},
	} {
		t.Run("", compareSpan(tt.want, MakeStrictBounds(tt.in, cmp.Compare, Next)))
	}
}

func TestReverse(t *testing.T) {
	want := s(b('A', 'Z'), b('a', 'z'))
	got := s(b('a', 'z'), b('A', 'Z'))
	t.Run("", compareSpan(want, got))
}

func fold(r Span[rune]) Span[rune] {
	rb := r.Bounds()
	for _, b := range rb {
		lo, hi := folded(b.Lo.Value, b.Hi.Value)
		r = r.UnionBound(NewBoundII(lo, hi))
	}

	return r
}

func folded(lo, hi rune) (_, _ rune) {
	lof, hif := unicode.SimpleFold(lo), unicode.SimpleFold(hi)
	if lo == lof || hi == hif {
		return lo, hi
	}

	return lof, hif
}

func compareSpan[T comparable](want, got Span[T]) func(*testing.T) {
	return func(t *testing.T) { t.Helper(); requireEqualSpan(t, want, got) }
}

func requireEqualSpan[T comparable](t *testing.T, want, got Span[T]) {
	t.Helper()
	if !IsEqual(want, got) {
		t.Logf("Not equal: \n"+
			"expected: %v\n"+
			"actual  : %v", want, got)
		t.FailNow()
	}
}

func compare[T comparable](want, got T) func(*testing.T) {
	return func(t *testing.T) { t.Helper(); requireEqual(t, want, got) }
}

func requireEqual[T comparable](t *testing.T, want, got T) {
	t.Helper()
	if want != got {
		t.Logf("Not equal: \n"+
			"expected: %v\n"+
			"actual  : %v", want, got)
		t.FailNow()
	}
}
