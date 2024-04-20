// Copyright (c) 2020-2024 Richard Cooper
//
// This file is a part of quenbyako/ext package.
// See https://github.com/quenbyako/ext/blob/master/LICENSE for details

package span_test

import (
	"cmp"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/quenbyako/ext/slices"

	. "github.com/quenbyako/ext/span"
)

type ttype = float64

func bl(b ...Bound[ttype]) []Bound[ttype] { return b }
func bxx(lo, hi ttype) Bound[ttype]       { return NewBoundXX(lo, hi) }
func bxi(lo, hi ttype) Bound[ttype]       { return NewBoundXI(lo, hi) }
func bix(lo, hi ttype) Bound[ttype]       { return NewBoundIX(lo, hi) }
func bii(lo, hi ttype) Bound[ttype]       { return NewBoundII(lo, hi) }

func bf(s string) Bound[ttype] {
	if b, err := ParseBound(s, func(s string) (ttype, error) { return strconv.ParseFloat(s, 10) }); err != nil {
		panic(err)
	} else {
		return b
	}
}

func blf(s string) []Bound[ttype] { return slices.Remap(strings.Split(s, " "), bf) }

func bi(s string) Bound[int] {
	if b, err := ParseBound(s, strconv.Atoi); err != nil {
		panic(err)
	} else {
		return b
	}
}

func bli(s string) []Bound[int] { return slices.Remap(strings.Split(s, " "), bi) }

func TestContains(t *testing.T) {
	for _, tt := range []struct {
		a, b Bound[ttype]
		want bool
	}{
		{bf("(0:3)"), bf("[1:2]"), true},
		{bf("(0:3)"), bf("[1:3]"), false},
		{bf("[1:2]"), bf("(1:2)"), true},
		{bf("(1:2)"), bf("[1:2]"), false},
		{bii(math.SmallestNonzeroFloat64, 1), bf("(0:1]"), false},
	} {
		t.Run("", func(t *testing.T) {
			if got := tt.a.Contains(cmp.Compare, tt.b); tt.want != got {
				t.Log(fmt.Sprintf("Not equal: \n"+
					"expected: %v\n"+
					"actual  : %v", tt.want, got))
				t.FailNow()
			}
		})
	}
}

func TestPosition(t *testing.T) {
	for _, tt := range []struct {
		a    Bound[ttype]
		v    ttype
		want int
	}{
		{bf("(0:1)"), 0, +1},
		{bf("(0:1]"), 0, +1},
		{bf("[0:1)"), 0, 0},
		{bf("[0:1]"), 0, 0},
		{bii(-1, -math.SmallestNonzeroFloat64), 0, -1},
	} {
		t.Run("", compare(tt.want, tt.a.Position(cmp.Compare, tt.v)))
	}
}

func TestUnion(t *testing.T) {
	for _, tt := range []struct {
		a, b, want Bound[ttype]
		wantOK     bool
	}{
		{bf("(0:1]"), bf("[1:2)"), bf("(0:2)"), true},
		{bf("(0:1)"), bf("[1:2)"), bf("(0:2)"), true},
		{bf("[-1:0]"), bii(math.SmallestNonzeroFloat64, 1), bf("[-1:0]"), false},
	} {
		t.Run("", func(t *testing.T) {
			got, gotOK := UnionBounds(nil, cmp.Compare, tt.a, tt.b)
			requireEqualBound(t, tt.want, got)
			requireEqual(t, tt.wantOK, gotOK)
		})
	}
}

func TestDifference(t *testing.T) {
	for _, tt := range []struct {
		a, b Bound[ttype]
		want []Bound[ttype]
	}{
		{bf("[1:3]"), bf("(1:2)"), blf("[1:1] [2:3]")}, // 0
		{bf("[1:3]"), bf("(1:2]"), blf("[1:1] (2:3]")}, // 1
		{bf("[1:2]"), bf("[1:3]"), nil},                // 2
		{bf("[1:4]"), bf("[2:3]"), blf("[1:2) (3:4]")}, // 3
		{bf("[1:4]"), bf("(1:4)"), blf("[1:1] [4:4]")}, // 4
		{bf("[1:3]"), bf("(1:2]"), blf("[1:1] (2:3]")}, // 5
		{bf("[1:3]"), bf("(0:4)"), nil},                // 6
		{bf("[1:3]"), bf("[0:2]"), blf("(2:3]")},       // 7
		{bf("[1:3]"), bf("[0:4]"), nil},                // 8
		{bf("(1:3)"), bf("[2:4]"), blf("(1:2)")},       // 9
		{bf("(1:3)"), bf("(2:4)"), blf("(1:2]")},       // 10
		{bf("(1:3]"), bf("(2:4)"), blf("(1:2]")},       // 11
		{bf("[1:3)"), bf("(2:4)"), blf("[1:2]")},       // 12
		{bf("[1:3]"), bf("(0:1]"), blf("(1:3]")},       // 13
		{bf("[1:3]"), bf("[1:1]"), blf("(1:3]")},       // 14
		{bf("(1:3)"), bf("[1:1]"), blf("(1:3)")},       // 15
		{bf("[1:1]"), bf("(1:3)"), blf("[1:1]")},       // 16
		{bf("[1:1]"), bf("[1:3]"), nil},                // 17
		{bf("[1:1]"), bf("[1:1]"), nil},                // 18
	} {
		t.Run("", compareBounds(tt.want, tt.a.Difference(cmp.Compare, tt.b)))
	}
}

func TestOverlaps(t *testing.T) {
	for _, tt := range []struct {
		a, b Bound[ttype]
		want bool
	}{
		{bf("(1:2)"), bf("(2:3)"), false},
		{bf("(1:2]"), bf("(2:3)"), false},
		{bf("(1:2)"), bf("(2:3]"), false},
		{bf("(1:2]"), bf("[2:3)"), true},
		{bxi(-1, 0), bix(math.SmallestNonzeroFloat64, 1), false},
		{bf("(1:3)"), bf("(2:4)"), true},
		{bf("(2:4)"), bf("(1:3)"), true},
		{bf("(1:4)"), bf("(2:3)"), true},
		{bf("(2:3)"), bf("(1:4)"), true},
		{bf("(0:1)"), bf("[1:1]"), false},
		{bf("(0:1]"), bf("[1:1]"), true},
	} {
		t.Run("", func(t *testing.T) {
			if got := tt.a.Overlaps(cmp.Compare, tt.b); tt.want != got {
				t.Log(fmt.Sprintf("Not equal: \n"+
					"expected: %v\n"+
					"actual  : %v", tt.want, got))
				t.FailNow()
			}
		})
	}
}

func requireEqualBounds[T comparable](t *testing.T, want, got []Bound[T]) {
	if !slices.EqualFunc(want, got, IsBoundEqual[T]) {
		t.Log(fmt.Sprintf("Not equal: \n"+
			"expected: %v\n"+
			"actual  : %v", strBounds(want), strBounds(got)))
		t.FailNow()
	}
}

func compareBounds[T comparable](want, got []Bound[T]) func(*testing.T) {
	return func(t *testing.T) { requireEqualBounds(t, want, got) }
}

func strBounds[T any](b []Bound[T]) string {
	if len(b) == 0 {
		return "<nil>"
	}
	parts := make([]string, len(b))
	for i, item := range b {
		parts[i] = item.String()
	}
	return strings.Join(parts, " ")
}

func requireEqualBound[T comparable](t *testing.T, want, got Bound[T]) {
	t.Helper()
	if !IsBoundEqual(want, got) {
		t.Log(fmt.Sprintf("Not equal: \n"+
			"expected: %v\n"+
			"actual  : %v", want, got))
		t.FailNow()
	}
}

func compareBound[T comparable](want, got Bound[T]) func(*testing.T) {
	return func(t *testing.T) { t.Helper(); requireEqualBound(t, want, got) }
}
