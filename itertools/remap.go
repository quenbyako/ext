package itertools

// Remap transforms values with `f` function from [iter.Seq] to [iter.Seq].
func Remap[E, J any](i iterSeq[E], f func(E) J) iterSeq[J] {
	return func(yield func(J) bool) {
		i(func(v E) bool { return yield(f(v)) })
	}
}

// RemapPairs transforms values with `f` function from [iter.Seq2] to [iter.Seq2].
func RemapPairs[K1, V1, K2, V2 any](i iterSeq2[K1, V1], f func(K1, V1) (K2, V2)) iterSeq2[K2, V2] {
	return func(yield func(K2, V2) bool) {
		i(func(k1 K1, v1 V1) bool { return yield(f(k1, v1)) })
	}
}

// function remaps [iter.Seq] to [iter.Seq2].
func SeqToPairs[E, K, V any](i iterSeq[E], f func(E) (K, V)) iterSeq2[K, V] {
	return func(yield func(K, V) bool) {
		i(func(v E) bool { return yield(f(v)) })
	}
}

// function remaps [iter.Seq2] to [iter.Seq].
func PairsToSeq[K, V, E any](i iterSeq2[K, V], f func(K, V) E) iterSeq[E] {
	return func(yield func(E) bool) {
		i(func(k K, v V) bool { return yield(f(k, v)) })
	}
}

// function remaps [iter.Seq2][K, V] to [iter.Seq][K].
func Keys[K, V any](i iterSeq2[K, V]) iterSeq[K] {
	return PairsToSeq(i, func(k K, _ V) K { return k })
}

// function remaps [iter.Seq2][K, V] to [iter.Seq][V].
func Values[K, V any](i iterSeq2[K, V]) iterSeq[V] {
	return PairsToSeq(i, func(_ K, v V) V { return v })
}

// Zip combines two sequences [iter.Seq][K] and [iter.Seq][V] into a single
// [iter.Seq2][K, V].
func Zip[K, V any](keys iterSeq[K], values iterSeq[V]) iterSeq2[K, V] {
	return func(yield func(K, V) bool) {
		var kList []K
		keys(func(k K) bool {
			kList = append(kList, k)
			return true
		})

		idx := 0
		values(func(v V) bool {
			if idx >= len(kList) {
				return false
			}
			if !yield(kList[idx], v) {
				return false
			}
			idx++
			return true
		})
	}
}
