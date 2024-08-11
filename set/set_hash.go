package set

import "hash"

type Hashable interface {
	Sum64() uint64
}

type setAny[T Hashable] map[uint64]T

var _ Set[hash.Hash64] = setAny[hash.Hash64](nil)

func newAnyNonTS[T Hashable](items ...T) Set[T] { return make(setAny[T], len(items)).Add(items...) }

// Add includes the specified items (one or more) to the set. The underlying
// Set s is modified. If passed nothing it silently returns.
func (s setAny[T]) Add(items ...T) Set[T] {
	for _, item := range items {
		s[item.Sum64()] = item
	}

	return s
}

// Remove deletes the specified items from the set.  The underlying Set s is
// modified. If passed nothing it silently returns.
func (s setAny[T]) Del(items ...T) Set[T] {
	for _, item := range items {
		delete(s, item.Sum64())
	}

	return s
}

func (s setAny[T]) Has(item T) bool {
	_, ok := s[item.Sum64()]

	return ok
}

func (s setAny[T]) Len() int { return len(s) }

func (s setAny[T]) Each(f func(item T) bool) {
	for _, item := range s {
		if !f(item) {
			break
		}
	}
}
