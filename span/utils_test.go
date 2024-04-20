// Copyright (c) 2020-2024 Richard Cooper
//
// This file is a part of quenbyako/ext package.
// See https://github.com/quenbyako/ext/blob/master/LICENSE for details

package span

import (
	"cmp"
	"math"
)

func S64(a ...Bound[float64]) Span[float64] {
	return span[float64]{
		next:   func(f float64) float64 { return math.Nextafter(f, math.Inf(+1)) },
		cmp:    cmp.Compare[float64],
		bounds: a,
	}
}

func Si(a ...Bound[int]) Span[int] {
	return span[int]{
		next:   func(f int) int { return f + 1 },
		cmp:    cmp.Compare[int],
		bounds: a,
	}
}

func Sr(a ...Bound[rune]) Span[rune] {
	return span[rune]{
		next:   func(f rune) rune { return f + 1 },
		cmp:    cmp.Compare[rune],
		bounds: a,
	}
}
