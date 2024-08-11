package set

import (
	"fmt"
	"hash"
	"sync"
	"sync/atomic"
)

type setts[T Hashable] struct {
	m   sync.Map // map[T]struct{}
	len atomic.Uint64
}

var _ Set[hash.Hash64] = (*setts[hash.Hash64])(nil)

//nolint:exhaustruct
func newTS[T Hashable](items ...T) Set[T] { return (&setts[T]{}).Add(items...) }

// Add includes the specified items (one or more) to the set. The underlying
// Set s is modified. If passed nothing it silently returns.
func (s *setts[T]) Add(items ...T) Set[T] {
	for _, item := range items {
		s.m.Store(item, struct{}{})
	}

	s.len.Add(uint64(len(items)))

	return s
}

// Remove deletes the specified items from the set.  The underlying Set s is
// modified. If passed nothing it silently returns.
func (s *setts[T]) Del(items ...T) Set[T] {
	for _, item := range items {
		s.m.Delete(item)
	}

	s.len.Add(uint64(-len(items)))

	return s
}

// Has looks for the existence of items passed. It returns false if nothing is
// passed. For multiple items it returns true only if all of  the items exist.
func (s *setts[T]) Has(item T) bool {
	_, ok := s.m.Load(item)

	return ok
}

// Len returns the number of items in the set.
//
// This method is not safe for concurrent use.
func (s *setts[T]) Len() int { return s.dirtyLen() }

// SafeLen returns the number of items in the set.
//
// Unlike [setts.Len], method is safe for concurrent use.
func (s *setts[T]) SafeLen() int { return s.safeLen() }

func (s *setts[T]) dirtyLen() int { return int(s.len.Load()) }
func (s *setts[T]) safeLen() int {
	var length int

	s.m.Range(func(_, _ interface{}) bool {
		length++

		return true
	})

	return length
}

//nolint:forcetypeassert // guaranteed
func (s *setts[T]) Each(yield func(T) bool) {
	s.m.Range(func(key, _ interface{}) bool { return yield(key.(T)) })
}

func (s *setts[T]) Format(f fmt.State, verb rune) {
	fmtValue := fmt.FormatString(f, verb)

	fmt.Fprintf(f, "set[")
	s.Each(func(b T) bool {
		fmt.Fprintf(f, fmtValue, b)

		return true
	})
}
