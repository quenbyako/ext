package itertools

import (
	"reflect"
	"testing"
)

func collect[T any](seq iterSeq[T]) []T {
	var res []T
	seq(func(v T) bool {
		res = append(res, v)
		return true
	})
	return res
}

func collectPairs[K, V any](seq iterSeq2[K, V]) ([]K, []V) {
	var keys []K
	var values []V
	seq(func(k K, v V) bool {
		keys = append(keys, k)
		values = append(values, v)
		return true
	})
	return keys, values
}

func seqOf[T any](vals ...T) iterSeq[T] {
	return func(yield func(T) bool) {
		for _, v := range vals {
			if !yield(v) {
				return
			}
		}
	}
}

func pairsOf[K, V any](keys []K, vals []V) iterSeq2[K, V] {
	return func(yield func(K, V) bool) {
		for i := 0; i < len(keys) && i < len(vals); i++ {
			if !yield(keys[i], vals[i]) {
				return
			}
		}
	}
}

func TestRemap(t *testing.T) {
	seq := seqOf(1, 2, 3)
	mapped := Remap(seq, func(v int) int { return v * 2 })
	got := collect(mapped)
	want := []int{2, 4, 6}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestRemapPairs(t *testing.T) {
	seq := pairsOf([]int{1, 2}, []string{"a", "b"})
	mapped := RemapPairs(seq, func(k int, v string) (int, string) {
		return k * 10, v + v
	})
	gotK, gotV := collectPairs(mapped)
	wantK, wantV := []int{10, 20}, []string{"aa", "bb"}
	if !reflect.DeepEqual(gotK, wantK) || !reflect.DeepEqual(gotV, wantV) {
		t.Errorf("got (%v, %v), want (%v, %v)", gotK, gotV, wantK, wantV)
	}
}

func TestSeqToPairs(t *testing.T) {
	seq := seqOf(1, 2)
	pairs := SeqToPairs(seq, func(v int) (int, int) { return v, v * 10 })
	gotK, gotV := collectPairs(pairs)
	wantK, wantV := []int{1, 2}, []int{10, 20}
	if !reflect.DeepEqual(gotK, wantK) || !reflect.DeepEqual(gotV, wantV) {
		t.Errorf("got (%v, %v), want (%v, %v)", gotK, gotV, wantK, wantV)
	}
}

func TestPairsToSeq(t *testing.T) {
	pairs := pairsOf([]int{1, 2}, []int{10, 20})
	seq := PairsToSeq(pairs, func(k, v int) int { return k + v })
	got := collect(seq)
	want := []int{11, 22}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestKeysAndValues(t *testing.T) {
	pairs := pairsOf([]int{1, 2}, []string{"a", "b"})
	keys := collect(Keys(pairs))
	values := collect(Values(pairs))

	if !reflect.DeepEqual(keys, []int{1, 2}) {
		t.Errorf("keys: got %v, want [1, 2]", keys)
	}
	if !reflect.DeepEqual(values, []string{"a", "b"}) {
		t.Errorf("values: got %v, want [a, b]", values)
	}
}

func TestZip(t *testing.T) {
	t.Run("EqualLength", func(t *testing.T) {
		keys := seqOf(1, 2, 3)
		values := seqOf("a", "b", "c")
		gotK, gotV := collectPairs(Zip(keys, values))
		if !reflect.DeepEqual(gotK, []int{1, 2, 3}) || !reflect.DeepEqual(gotV, []string{"a", "b", "c"}) {
			t.Errorf("got (%v, %v)", gotK, gotV)
		}
	})

	t.Run("KeysShorter", func(t *testing.T) {
		keys := seqOf(1, 2)
		values := seqOf("a", "b", "c")
		gotK, gotV := collectPairs(Zip(keys, values))
		if !reflect.DeepEqual(gotK, []int{1, 2}) || !reflect.DeepEqual(gotV, []string{"a", "b"}) {
			t.Errorf("got (%v, %v)", gotK, gotV)
		}
	})

	t.Run("ValuesShorter", func(t *testing.T) {
		keys := seqOf(1, 2, 3)
		values := seqOf("a", "b")
		gotK, gotV := collectPairs(Zip(keys, values))
		if !reflect.DeepEqual(gotK, []int{1, 2}) || !reflect.DeepEqual(gotV, []string{"a", "b"}) {
			t.Errorf("got (%v, %v)", gotK, gotV)
		}
	})

	t.Run("EarlyTermination", func(t *testing.T) {
		keys := seqOf(1, 2, 3)
		values := seqOf("a", "b", "c")
		zipped := Zip(keys, values)
		var gotK []int
		zipped(func(k int, v string) bool {
			gotK = append(gotK, k)
			return k < 2
		})
		if !reflect.DeepEqual(gotK, []int{1, 2}) {
			t.Errorf("got %v, want [1, 2]", gotK)
		}
	})
}
