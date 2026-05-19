//go:build !go1.23

package itertools

type (
	iterSeq2[K, V any] func(yield func(K, V) bool)
	iterSeq[K any]     func(yield func(K) bool)
)
