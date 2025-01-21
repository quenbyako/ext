package itertools

import "iter"

// Remap transforms values with `f` function from [iter.Seq] to [iter.Seq].
func Remap[E, J any](i iter.Seq[E], f func(E) J) iter.Seq[J] {
	return func(yield func(J) bool) {
		i(func(v E) bool { return yield(f(v)) })
	}
}

// RemapPairs transforms values with `f` function from [iter.Seq2] to [iter.Seq2].
func RemapPairs[K1, V1, K2, V2 any](i iter.Seq2[K1, V1], f func(K1, V1) (K2, V2)) iter.Seq2[K2, V2] {
	return func(yield func(K2, V2) bool) {
		i(func(k1 K1, v1 V1) bool { return yield(f(k1, v1)) })
	}
}

// function remaps [iter.Seq] to [iter.Seq2].
func SeqToPairs[E, K, V any](i iter.Seq[E], f func(E) (K, V)) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		i(func(v E) bool { return yield(f(v)) })
	}
}

// function remaps [iter.Seq2] to [iter.Seq].
func PairsToSeq[K, V, E any](i iter.Seq2[K, V], f func(K, V) E) iter.Seq[E] {
	return func(yield func(E) bool) {
		i(func(k K, v V) bool { return yield(f(k, v)) })
	}
}

// function remaps [iter.Seq2][K, V] to [iter.Seq][K].
func Keys[K, V any](i iter.Seq2[K, V]) iter.Seq[K] {
	return PairsToSeq(i, func(k K, _ V) K { return k })
}

// function remaps [iter.Seq2][K, V] to [iter.Seq][V].
func Values[K, V any](i iter.Seq2[K, V]) iter.Seq[V] {
	return PairsToSeq(i, func(_ K, v V) V { return v })
}

// Zip combines two sequences [iter.Seq][K] and [iter.Seq][V] into a single
// [iter.Seq2][K, V].
func Zip[K, V any](keys iter.Seq[K], values iter.Seq[V]) iter.Seq2[K, V] {
	// converting to pull iterators
	nextKeys, stopKeys := iter.Pull(keys)
	nextValues, stopValues := iter.Pull(values)

	return func(yield func(K, V) bool) {
		defer stopKeys()
		defer stopValues()

		for {
			// Getting next key and value
			key, okKey := nextKeys()
			value, okValue := nextValues()

			// if any of the iterators is exhausted, we are stopping
			if !okKey || !okValue {
				break
			}

			// Yielding key and value
			if !yield(key, value) {
				break
			}
		}
	}
}
