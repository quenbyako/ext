//go:build !go1.22

package span

import "cmp"

func compare[T cmp.Ordered](x, y T) int {
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
