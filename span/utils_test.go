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
		next:   NextInt[int],
		cmp:    cmp.Compare[int],
		bounds: a,
	}
}

func Sr(a ...Bound[rune]) Span[rune] {
	return span[rune]{
		next:   NextInt[rune],
		cmp:    cmp.Compare[rune],
		bounds: a,
	}
}


func Compare[T cmp.Ordered](x, y T) int {
	switch xNaN, yNaN := x != x, y != y; {
	case xNaN && yNaN:
		return 0
	case xNaN:
		return -1
	case yNaN:
		return +1
	case x < y:
		return -1
	case x > y:
		return +1
	default:
		return 0
	}
}
