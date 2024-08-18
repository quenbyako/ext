//go:build go1.22

package span

import "cmp"

func compare[T cmp.Ordered](a, b T) int { return cmp.Compare(a, b) }
