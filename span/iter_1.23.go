//go:build go1.23

package span

import "iter"

type (
	iterSeq2[K, V any] = iter.Seq2[K, V]
	iterSeq[K any]     = iter.Seq[K]
)
