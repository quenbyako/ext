package set

import (
	"sync"
)

// setm defines a thread safe set data structure.
type setm[T any] struct {
	s            Set[T]
	sync.RWMutex // we name it because we don't want to expose it
}

var _ interface {
	rwLocker
	Set[int]
} = (*setm[int])(nil)

// New creates and initialize a new Set. It's accept a variable number of
// arguments to populate the initial set. If nothing passed a Set with zero
// size is created.
func wrapMutex[T any](s Set[T]) Set[T] { return &setm[T]{s: s} }

type rwLocker interface {
	RLock()
	RUnlock()
}

// Add includes the specified items (one or more) to the set. The underlying
// Set s is modified. If passed nothing it silently returns.
func (s *setm[T]) Add(items ...T) Set[T] {
	s.Lock()
	defer s.Unlock()
	s.s = s.s.Add(items...)

	return s
}

// Remove deletes the specified items from the set.  The underlying Set s is
// modified. If passed nothing it silently returns.
func (s *setm[T]) Remove(items ...T) Set[T] {
	s.Lock()
	defer s.Unlock()
	s.s = s.s.Remove(items...)

	return s
}

func (s *setm[T]) Separate(t Set[T]) Set[T] {
	s.Lock()
	defer s.Unlock()
	s.s = s.s.Separate(t)

	return s
}

func (s *setm[T]) String() string {
	s.RLock()
	defer s.RUnlock()
	return s.s.String()
}

// Pop  deletes and return an item from the set. The underlying Set s is
// modified. If set is empty, nil is returned.
func (s *setm[T]) Pop() (T, bool) {
	s.Lock()
	defer s.Unlock()
	return s.s.Pop()
}

// Has looks for the existence of items passed. It returns false if nothing is
// passed. For multiple items it returns true only if all of  the items exist.
func (s *setm[T]) Has(items ...T) bool {
	s.RLock()
	defer s.RUnlock()
	return s.s.Has(items...)
}

// Size returns the number of items in a set.
func (s *setm[T]) Size() int {
	s.RLock()
	defer s.RUnlock()
	return s.s.Size()
}

// Clear removes all items from the set.
func (s *setm[T]) Clear() {
	s.Lock()
	defer s.Unlock()
	s.s.Clear()
}

// IsEqual test whether s and t are the same in size and have the same items.
func (s *setm[T]) IsEqual(t Set[T]) bool {
	s.RLock()
	defer s.RUnlock()
	return s.s.IsEqual(t)
}

func (s *setm[T]) IsEmpty() bool {
	s.RLock()
	defer s.RUnlock()
	return s.s.IsEmpty()
}

// IsSubset tests whether t is a subset of s.
func (s *setm[T]) IsSubset(t Set[T]) bool {
	s.RLock()
	defer s.RUnlock()
	return s.s.IsSubset(t)
}

func (s *setm[T]) IsSuperset(t Set[T]) bool {
	s.RLock()
	defer s.RUnlock()
	return s.s.IsSuperset(t)
}

func (s *setm[T]) Each(f func(item T) bool) bool {
	s.RLock()
	defer s.RUnlock()
	return s.s.Each(f)
}

// List returns a slice of all items.
func (s *setm[T]) List() []T {
	s.RLock()
	defer s.RUnlock()
	return s.s.List()
}

func (s *setm[T]) Copy() Set[T] {
	s.RLock()
	defer s.RUnlock()
	return wrapMutex(s.s.Copy())
}

func (s *setm[T]) Merge(t Set[T]) Set[T] {
	s.Lock()
	defer s.Unlock()
	s.s = s.s.Merge(t)

	return s
}
