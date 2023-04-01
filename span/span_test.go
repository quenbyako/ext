package span_test

import (
	"fmt"
	"reflect"
	"testing"
	"unicode"

	. "github.com/quenbyako/ext/span"
)

func s(b ...Bound[rune]) Span[rune] { return New(b...) }
func b(lo, hi rune) Bound[rune]     { return NewBound(lo, hi) }
func r(r rune) Bound[rune]          { return NewBound(r, r) }

func TestIn(t *testing.T) {
	for _, tt := range []struct {
		a    Span[rune]
		r    rune
		want bool
	}{
		{s(b('a', 'a')), 'a', true},
		{s(b('a', 'z')), 'o', true},
		{s(b('a', 'z')), '0', false},
		{s(b('0', '9'), b('a', 'z')), '1', true},
		{s(b('0', '9'), b('a', 'z')), 'b', true},
		{s(b('0', '9'), b('a', 'z')), '@', false},
	} {
		t.Run("", compare(tt.want, tt.a.In(tt.r)))
	}
}

func TestMerge(t *testing.T) {
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
		t.Run("", compare(tt.want, tt.a.Merge(tt.b)))
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
		t.Run("", compare(tt.want, fold(tt.in)))
	}
}

func fold(r Span[rune]) Span[rune] {
	rb := r.Bounds()
	newItems := make([]Bound[rune], len(rb))
	for i, b := range rb {
		lo, hi := folded(b.Lo(), b.Hi())
		newItems[i] = NewBound(lo, hi)
	}

	return r.Merge(New(newItems...))
}

func folded(lo, hi rune) (_, _ rune) {
	lof, hif := unicode.SimpleFold(lo), unicode.SimpleFold(hi)
	if lo == lof || hi == hif {
		return lo, hi
	}

	return lof, hif
}

func compare(want, got any) func(*testing.T) {
	return func(t *testing.T) {
		if !reflect.DeepEqual(want, got) {
			t.Log(fmt.Sprintf("Not equal: \n"+
				"expected: %v\n"+
				"actual  : %v", want, got))
			t.FailNow()
		}
	}
}
