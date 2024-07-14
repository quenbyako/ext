// Copyright (c) 2020-2024 Richard Cooper
//
// This file is a part of quenbyako/ext package.
// See https://github.com/quenbyako/ext/blob/master/LICENSE for details

package span_test

import (
	"cmp"
	"errors"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/quenbyako/ext/slices"

	. "github.com/quenbyako/ext/span"
)

type ttype = float64

func bxi(lo, hi ttype) Bound[ttype] { return NewBoundXI(lo, hi) }
func bix(lo, hi ttype) Bound[ttype] { return NewBoundIX(lo, hi) }
func bii(lo, hi ttype) Bound[ttype] { return NewBoundII(lo, hi) }

func sr(b ...string) Span[rune] {
	bounds := make([]Bound[rune], len(b))
	for i, item := range b {
		bounds[i] = br(item)
	}

	return New(Next[rune], cmp.Compare[rune], bounds...)
}

func br(s string) Bound[rune] {
	if b, err := ParseBound(s, func(s string) (rune, error) {
		if r := []rune(s); len(r) == 1 {
			return r[0], nil
		}
		return 0, errors.New("invalid value, should be only one rune")
	}); err != nil {
		panic(err)
	} else {
		return b
	}
}

func bf(s string) Bound[ttype] {
	if b, err := ParseBound(s, func(s string) (ttype, error) { return strconv.ParseFloat(s, 64) }); err != nil {
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
				t.Logf("Not equal: \n"+
					"expected: %v\n"+
					"actual  : %v", tt.want, got)
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
		name string
		a, b Bound[ttype]
		want []Bound[ttype]
	}{
		{"#00", bf("[1:3]"), bf("(1:2)"), blf("[1:1] [2:3]")},
		{"#01", bf("[1:3]"), bf("(1:2]"), blf("[1:1] (2:3]")},
		{"#02", bf("[1:2]"), bf("[1:3]"), nil},
		{"#03", bf("[1:4]"), bf("[2:3]"), blf("[1:2) (3:4]")},
		{"#04", bf("[1:4]"), bf("(1:4)"), blf("[1:1] [4:4]")},
		{"#05", bf("[1:3]"), bf("(1:2]"), blf("[1:1] (2:3]")},
		{"#06", bf("[1:3]"), bf("(0:4)"), nil},
		{"#07", bf("[1:3]"), bf("[0:2]"), blf("(2:3]")},
		{"#08", bf("[1:3]"), bf("[0:4]"), nil},
		{"#09", bf("(1:3)"), bf("[2:4]"), blf("(1:2)")},
		{"#10", bf("(1:3)"), bf("(2:4)"), blf("(1:2]")},
		{"#11", bf("(1:3]"), bf("(2:4)"), blf("(1:2]")},
		{"#12", bf("[1:3)"), bf("(2:4)"), blf("[1:2]")},
		{"#13", bf("[1:3]"), bf("(0:1]"), blf("(1:3]")},
		{"#14", bf("[1:3]"), bf("[1:1]"), blf("(1:3]")},
		{"#15", bf("(1:3)"), bf("[1:1]"), blf("(1:3)")},
		{"#16", bf("[1:1]"), bf("(1:3)"), blf("[1:1]")},
		{"#17", bf("[1:1]"), bf("[1:3]"), nil},
		{"#18", bf("[1:1]"), bf("[1:1]"), nil},
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
				t.Logf("Not equal: \n"+
					"expected: %v\n"+
					"actual  : %v", tt.want, got)
				t.FailNow()
			}
		})
	}
}

func requireEqualBounds[T comparable](t *testing.T, want, got []Bound[T]) {
	if !slices.Equal(want, got) {
		t.Logf("Not equal: \n"+
			"expected: %v\n"+
			"actual  : %v", strBounds(want), strBounds(got))
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
	if want != got {
		t.Logf("Not equal: \n"+
			"expected: %v\n"+
			"actual  : %v", want, got)
		t.FailNow()
	}
}
