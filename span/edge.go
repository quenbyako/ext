// Copyright (c) 2020-2024 Richard Cooper
//
// This file is a part of quenbyako/ext package.
// See https://github.com/quenbyako/ext/blob/master/LICENSE for details

package span

type Edge[T any] struct {
	Value    T
	Included bool
}

func newEdge[T any](v T, i bool) Edge[T] { return Edge[T]{Value: v, Included: i} }

func IsEdgeNear[T any](next nextFunc[T], cmp compareFunc[T], lower, higher Edge[T]) bool {
	// cases:
	// * [1:2] [2:3]
	// * [1:2) [2:3]
	// * [1:2] (2:3]
	// BUT NOT:
	// * [1:2) (2:3] // 2 is excluded
	//
	// checks ALWAYS.
	if cmp(lower.Value, higher.Value) == 0 && (lower.Included || higher.Included) {
		return true
	}

	// cases:
	// * [1:2] [3:4] // there are no values between 2 and 3
	// BUT NOT:
	// * [1:2] (3:4] // 3 is not in bound
	// * [1:2) [3:4] // 2 is not in bound
	// * [1:2) (3:4] // missed 2 and 3
	//
	// checks ONLY if next function is provided
	if next != nil && lower.Included && higher.Included && cmp(next(lower.Value, higher.Value), higher.Value) >= 0 {
		return true
	}

	return false
}

func minEdge[T any](a, b Edge[T], cmp func(T, T) int) Edge[T] {
	switch compared := cmp(a.Value, b.Value); {
	case compared > 0:
		return b
	case compared < 0:
		return a
	default: // a == b
		return newEdge(a.Value, a.Included || b.Included)
	}
}

func maxEdge[T any](a, b Edge[T], cmp func(T, T) int) Edge[T] {
	switch compared := cmp(a.Value, b.Value); {
	case compared > 0:
		return a
	case compared < 0:
		return b
	default: // a == b
		return newEdge(a.Value, a.Included || b.Included)
	}
}
