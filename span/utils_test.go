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
		next:   math.Nextafter,
		cmp:    cmp.Compare[float64],
		bounds: a,
	}
}

func Si(a ...Bound[int]) Span[int] {
	return span[int]{
		next:   Next[int],
		cmp:    cmp.Compare[int],
		bounds: a,
	}
}

func Sr(a ...Bound[rune]) Span[rune] {
	return span[rune]{
		next:   Next[rune],
		cmp:    cmp.Compare[rune],
		bounds: a,
	}
}

func Next[T int | rune](v, t T) T {
	switch {
	case v == t:
		return v
	case v < t:
		return v + 1
	default:
		return v - 1
	}
}
