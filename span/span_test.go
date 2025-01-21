// Copyright (c) 2020-2024 Richard Cooper
//
// This file is a part of quenbyako/ext package.
// See https://github.com/quenbyako/ext/blob/master/LICENSE for details

package span_test

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"testing"
	"unicode"

	. "github.com/quenbyako/ext/span"
)

func s(b ...Bound[rune]) Span[rune]      { return NewRune(b...) }
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
			_ = tt // requireEqual(t, tt.want, tt.a.Search(tt.r))
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

func TestSubtractSpans(t *testing.T) {
	for _, tt := range []struct{ a, b, want Span[int] }{
		{Si(bli("[1:6]")...), Si(bli("[2:4]")...), Si(bli("[1:2) (4:6]")...)},
		{Si(bli("[1:6]")...), Si(bli("[2:2]")...), Si(bli("[1:2) (2:6]")...)},
		{Si(bli("[1:3) [4:6]")...), Si(bli("[3:4]")...), Si(bli("[1:3) (4:6]")...)},
	} {
		t.Run("", compareSpan(tt.want, tt.a.Subtract(tt.b)))
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
		{NewRune(Bound[rune]{Edge[rune]{Value: 1, Included: false}, Edge[rune]{Value: 2, Included: false}}), NewRune()},
	} {
		t.Run("", compareSpan(tt.want, MakeStrictBounds(tt.in, cmp.Compare[rune], NextInt[rune])))
	}
}

func TestSplit(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   []Span[rune]
		want map[Bound[rune]][]int
	}{{
		//  0: Split([1:3][6:7], [2:3][5:6]) {
		//        return {
		//            [1:2) -> [0]
		//            [2:3] -> [0, 1]
		//            [5:6) -> [1]
		//            [6:6] -> [0, 1]
		//            (6:7] -> [0]
		//        }
		//     }
		name: "0",
		in:   []Span[rune]{sr("[1:3]", "[6:7]"), sr("[2:3]", "[5:6]")},
		want: map[Bound[rune]][]int{
			br("[1:2)"): {0},
			br("[2:3]"): {0, 1},
			br("[5:6)"): {1},
			br("[6:6]"): {0, 1},
			br("(6:7]"): {0},
		},
	}, {
		//  1: Split([1:5], [2:4], [3:6]) {
		//        return {
		//            [1:2) -> [0]
		//            [2:3) -> [0, 1]
		//            [3:4] -> [0, 1, 2]
		//            (4:5] -> [0, 2]
		//            (5:6] -> [2]
		//        }
		//     }
		name: "1",
		in:   []Span[rune]{sr("[1:5]"), sr("[2:4]"), sr("[3:6]")},
		want: map[Bound[rune]][]int{
			br("[1:2)"): {0},
			br("[2:3)"): {0, 1},
			br("[3:4]"): {0, 1, 2},
			br("(4:5]"): {0, 2},
			br("(5:6]"): {2},
		},
	}, {
		//  2: Split([1:3], [3:7)) {
		//        return {
		//            [1:3) -> [0]
		//            [3:3] -> [0, 1]
		//            (3:7) -> [1]
		//        }
		//     }
		name: "2-collide",
		in:   []Span[rune]{sr("[1:3]"), sr("[3:7]")},
		want: map[Bound[rune]][]int{
			br("[1:3)"): {0},
			br("[3:3]"): {0, 1},
			br("(3:7]"): {1},
		},
	}, {
		//  2: Split([1:3), (3:7)) {
		//        return {
		//            [1:3) -> [0]
		//            (3:7) -> [1]
		//        }
		//     }
		name: "2-none",
		in:   []Span[rune]{sr("[1:3)"), sr("(3:7]")},
		want: map[Bound[rune]][]int{
			br("[1:3)"): {0},
			br("(3:7]"): {1},
		},
	}, {
		//  2: Split([1:3), [3:7)) {
		//        return {
		//            [1:3] -> [0]
		//            (3:7) -> [1]
		//        }
		//     }
		name: "2-right-more",
		in:   []Span[rune]{sr("[1:3)"), sr("[3:7]")},
		want: map[Bound[rune]][]int{
			br("[1:3)"): {0},
			br("[3:7]"): {1},
		},
	}, {
		//  2: Split([1:3], (3:7)) {
		//        return {
		//            [1:3] -> [0]
		//            (3:7) -> [1]
		//        }
		//     }
		name: "2-left-more",
		in:   []Span[rune]{sr("[1:3]"), sr("(3:7]")},
		want: map[Bound[rune]][]int{
			br("[1:3]"): {0},
			br("(3:7]"): {1},
		},
	}, {
		//  3: Split([1:1], [2:2], [3:3]) {
		//        return {
		//            [1:1] -> [0]
		//            [2:2] -> [1]
		//            [3:3] -> [2]
		//        }
		//     }
		name: "3",

		in: []Span[rune]{sr("[1:1]"), sr("[2:2]"), sr("[3:3]")},
		want: map[Bound[rune]][]int{
			br("[1:1]"): {0},
			br("[2:2]"): {1},
			br("[3:3]"): {2},
		},
	}, {
		//  4: Split([1:2), (2:3], [3:5]) {
		//        return {
		//           [1:2) -> [0]
		//           (2:3) -> [1] // not exists
		//           [3:3] -> [1,2]
		//           (3:5] -> [2]
		//       }
		//     }
		name: "4",
		in:   []Span[rune]{sr("[1:2)"), sr("(2:3]"), sr("[3:5]")},
		want: map[Bound[rune]][]int{
			br("[1:2)"): {0},
			// br("(2:3)"): {1}, // not exists
			br("[3:3]"): {1, 2},
			br("(3:5]"): {2},
		},
	}, {
		//  5: Split([0:9], [0:0], [9:9]) {
		//        return {
		//            [0:0] -> [0, 1]
		//            (0:9) -> [0]
		//            [9:9] -> [0, 2]
		//        }
		//     }
		name: "5",
		in:   []Span[rune]{sr("[0:9]"), sr("[0:0]"), sr("[9:9]")},
		want: map[Bound[rune]][]int{
			br("[0:0]"): {0, 1},
			br("(0:9)"): {0},
			br("[9:9]"): {0, 2},
		},
	}, {
		// 6: Split() {
		//      return {}
		// }
		name: "6-none",
		in:   []Span[rune]{},
		want: map[Bound[rune]][]int{},
	}, {
		// 7: Split([1:2]) {
		//      return {[1:2] -> [0]}
		//}
		name: "7-single",
		in:   []Span[rune]{sr("[1:2]")},
		want: map[Bound[rune]][]int{
			br("[1:2]"): {0},
		},
	}, {
		//  8: Split([0:1], [1:2], [2:3], [3:4], [4:5]) {
		//        return {
		//            [0:1) -> [0]
		//            [1:1] -> [0, 1]
		//            (1:2) -> [1] // not exists
		//            [2:2] -> [1, 2]
		//            (2:3) -> [2] // not exists
		//            [3:3] -> [2, 3]
		//            (3:4) -> [3] // not exists
		//            [4:4] -> [3, 4]
		//            (4:5] -> [4]
		//        }
		//     }
		name: "8",
		in:   []Span[rune]{sr("[0:1]"), sr("[1:2]"), sr("[2:3]"), sr("[3:4]"), sr("[4:5]")},
		want: map[Bound[rune]][]int{
			br("[0:1)"): {0},
			br("[1:1]"): {0, 1},
			// br("(1:2)"): {1}, // not exists
			br("[2:2]"): {1, 2},
			// br("(2:3)"): {2}, // not exists
			br("[3:3]"): {2, 3},
			// br("(3:4)"): {3}, // not exists
			br("[4:4]"): {3, 4},
			br("(4:5]"): {4},
		},
	}, {
		//  9: Split([0:2), [1:2], [0:4], [2:3], (2:4]) {
		//        return {
		//            [0:1) -> [0, 2]
		//            [1:1] -> [0, 1, 2]
		//            (1:2) -> [0, 1, 2]
		//            [2:2] -> [1, 2, 3]
		//            (2:3) -> [2, 3, 4]
		//            [3:3] -> [2, 3, 4]
		//            (3:4] -> [2, 4]
		//        }
		//     }
		name: "9-all-together",
		in:   []Span[rune]{sr("[0:2)"), sr("[1:2]"), sr("[0:4]"), sr("[2:3]"), sr("(2:4]")},
		want: map[Bound[rune]][]int{
			br("[0:1)"): {0, 2},
			br("[1:2)"): {0, 1, 2},
			br("[2:2]"): {1, 2, 3},
			br("(2:3]"): {2, 3, 4},
			br("(3:4]"): {2, 4},
		},
	}, {
		//  10: Split([0:2), [1:2], [0:4], [2:3], [2:4]) {
		//        return {
		//            [0:1) -> [0, 2]
		//            [1:1] -> [0, 1, 2]
		//            (1:2) -> [0, 1, 2]
		//            [2:2] -> [1, 2, 3]
		//            (2:3) -> [2, 3, 4]
		//            [3:3] -> [2, 3, 4]
		//            (3:4] -> [2, 4]
		//        }
		//     }
		name: "10",
		in:   []Span[rune]{sr("[0:2)"), sr("[1:2]"), sr("[0:4]"), sr("[2:3]"), sr("[2:4]")},
		want: map[Bound[rune]][]int{
			br("[0:1)"): {0, 2},
			br("[1:2)"): {0, 1, 2},
			br("[2:2]"): {1, 2, 3, 4},
			br("(2:3]"): {2, 3, 4},
			br("(3:4]"): {2, 4},
		},
	}, {
		name: "11",
		in:   []Span[rune]{sr("(0:2)"), sr("[1:2]"), sr("(0:4]"), sr("(2:4]")},
		want: map[Bound[rune]][]int{
			// br("(0:1)"): {0, 2}, // not exists
			br("[1:2)"): {0, 1, 2},
			br("[2:2]"): {1, 2},
			br("(2:4]"): {2, 3},
		},
	}, {
		//  12: Split([0:0], (0:1]) {
		//        return {
		//            [0:0] -> [0]
		//            (0:1] -> [1]
		//        }
		//     }
		name: "12",
		in:   []Span[rune]{sr("[0:0]"), sr("(0:1]")},
		want: map[Bound[rune]][]int{
			br("[0:0]"): {0},
			br("(0:1]"): {1},
		},
	}, {
		//  13: Split([0:1], (0:1]) {
		//        return {
		//            [0:0] -> [0]
		//            (0:1] -> [0, 1]
		//        }
		//     }
		name: "13",
		in:   []Span[rune]{sr("[0:1]"), sr("(0:1]")},
		want: map[Bound[rune]][]int{
			br("[0:0]"): {0},
			br("(0:1]"): {0, 1},
		},
	}, {
		//  14: Split([0:2], (1:2]) {
		//        return {
		//            [0:1] -> [0]
		//            (1:2] -> [0, 1]
		//        }
		//     }
		name: "14",
		in:   []Span[rune]{sr("[0:2]"), sr("(1:2]")},
		want: map[Bound[rune]][]int{
			br("[0:1]"): {0},
			br("(1:2]"): {0, 1},
		},
	}, {
		//  15: Split([0:2], [1:2], (1:2]) {
		//        return {
		//            [0:1) -> [0]
		//            [1:1] -> [0, 1]
		//            (1:2] -> [0, 1, 2]
		//        }
		//     }
		name: "15",
		in:   []Span[rune]{sr("[0:2]"), sr("[1:2]"), sr("(1:2]")},
		want: map[Bound[rune]][]int{
			br("[0:1)"): {0},
			br("[1:1]"): {0, 1},
			br("(1:2]"): {0, 1, 2},
		},
	}, {
		//  16: Split([0:2], [1:1], (1:2]) {
		//        return {
		//            [0:1) -> [0]
		//            [1:1] -> [0, 1]
		//            (1:2] -> [0, 2]
		//        }
		//     }
		name: "16",
		in:   []Span[rune]{sr("[0:2]"), sr("[1:1]"), sr("(1:2]")},
		want: map[Bound[rune]][]int{
			br("[0:1)"): {0},
			br("[1:1]"): {0, 1},
			br("(1:2]"): {0, 2},
		},
	}, {
		//  17: Split([0:1), [0:2]) {
		//        return {
		//            [0:1) -> [0, 1]
		//            [1:2] -> [1]
		//        }
		//     }
		name: "17",
		in:   []Span[rune]{sr("[0:1)"), sr("[0:2]")},
		want: map[Bound[rune]][]int{
			br("[0:1)"): {0, 1},
			br("[1:2]"): {1},
		},
	}, {
		name: "18",
		in:   []Span[rune]{sr("[0:2]"), sr("[0:1)"), sr("(1:2]")},
		want: map[Bound[rune]][]int{
			br("[0:1)"): {0, 1},
			br("[1:1]"): {0},
			br("(1:2]"): {0, 2},
		},
	}, {
		name: "19",
		in:   []Span[rune]{sr("[0:1]"), sr("[0:1)"), sr("[0:2]")},
		want: map[Bound[rune]][]int{
			br("[0:1)"): {0, 1, 2},
			br("[1:1]"): {0, 2},
			br("(1:2]"): {2},
		},
	}, {
		name: "20",
		in:   []Span[rune]{sr("[0:1)"), sr("[1:2]"), sr("(1:2]")},
		want: map[Bound[rune]][]int{
			br("[0:1)"): {0},
			br("[1:1]"): {1},
			br("(1:2]"): {1, 2},
		},
	}, {
		name: "21",
		in:   []Span[rune]{sr("[0:1)"), sr("[1:2]")},
		want: map[Bound[rune]][]int{
			br("[0:1)"): {0},
			br("[1:2]"): {1},
		},
	}, {
		name: "22",
		in:   []Span[rune]{sr("[0:1)"), sr("[1:2]"), sr("[0:2]")},
		want: map[Bound[rune]][]int{
			br("[0:1)"): {0, 2},
			br("[1:2]"): {1, 2},
		},
	}} {
		t.Run(tt.name, func(t *testing.T) {
			got := Split(tt.in, cmp.Compare[rune], NextInt[rune])

			requireEqualMap(t, tt.want, got)
		})
	}
}

func TestReverse(t *testing.T) {
	want := s(b('A', 'Z'), b('a', 'z'))
	got := s(b('a', 'z'), b('A', 'Z'))
	t.Run("", compareSpan(want, got))
}

func fold(r Span[rune]) Span[rune] {
	for _, b := range r.Bounds() {
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

func requireEqualMap(t *testing.T, want, got map[Bound[rune]][]int) {
	t.Helper()

	if !maps.EqualFunc(want, got, slices.Equal[[]int]) {
		kwant := slices.Collect(maps.Keys(want))
		slices.SortFunc(kwant, cmpLoBound)
		kgot := slices.Collect(maps.Keys(got))
		slices.SortFunc(kgot, cmpLoBound)

		if !slices.Equal(kwant, kgot) {
			t.Logf("Not equal: \nExpected: %q\nActual:   %q", kwant, kgot)
			t.FailNow()
		}

		log := "Not equal: \nExpected: "
		for i, k := range kwant {
			if i > 0 {
				log += "          "
			}
			log += fmt.Sprintf("%q -> %v", k, want[k]) + "\n"
		}

		log += "\nActual:   "
		for i, k := range kgot {
			if i > 0 {
				log += "          "
			}
			log += fmt.Sprintf("%q -> %v", k, got[k]) + "\n"
		}

		t.Log(log)
		t.FailNow()
	}
}

func cmpLoBound(a, b Bound[rune]) int {
	ae := a.Lo
	be := b.Lo
	switch {
	case ae.Value > be.Value:
		return +1
	case ae.Value < be.Value:
		return -1

	// a == b
	case !ae.Included && be.Included: // (a, +Inf] > [a, +Inf]
		return +1
	case ae.Included && !be.Included: // [a, +Inf] < (a, +Inf]
		return -1

	// (a, +Inf] == (a, +Inf]
	// [a, +Inf] == [a, +Inf]
	default:
		return 0
	}
}
