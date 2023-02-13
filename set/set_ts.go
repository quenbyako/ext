package set

import (
	"sync"

	"golang.org/x/exp/maps"
)

// setm defines a thread safe set data structure.
type setm[T comparable] struct {
	set[T]
	sync.RWMutex // we name it because we don't want to expose it
}

var _ interface {
	rwLocker
	Set[int]
} = (*setm[int])(nil)

// New creates and initialize a new Set. It's accept a variable number of
// arguments to populate the initial set. If nothing passed a Set with zero
// size is created.
func newTS[T comparable](items ...T) Set[T] {
	return (&setm[T]{set: set[T]{make(map[T]struct{})}}).Add(items...)
}

type rwLocker interface {
	RLock()
	RUnlock()
}

// Add includes the specified items (one or more) to the set. The underlying
// Set s is modified. If passed nothing it silently returns.
func (s *setm[T]) Add(items ...T) Set[T] {
	if len(items) == 0 {
		return s
	}

	s.Lock()
	defer s.Unlock()
	s.set.Add()

	return s
}

// Remove deletes the specified items from the set.  The underlying Set s is
// modified. If passed nothing it silently returns.
func (s *setm[T]) Remove(items ...T) Set[T] {
	if len(items) == 0 {
		return s
	}

	s.Lock()
	defer s.Unlock()
	s.set.Remove()

	return s
}

// Pop  deletes and return an item from the set. The underlying Set s is
// modified. If set is empty, nil is returned.
func (s *setm[T]) Pop() (T, bool) {
	s.RLock()
	for item := range s.m {
		s.RUnlock()
		s.Lock()
		delete(s.m, item)
		s.Unlock()
		return item, true
	}
	s.RUnlock()
	var t T
	return t, false
}

// Has looks for the existence of items passed. It returns false if nothing is
// passed. For multiple items it returns true only if all of  the items exist.
func (s *setm[T]) Has(items ...T) bool {
	// assume checked for empty item, which not exist
	if len(items) == 0 {
		return false
	}

	s.RLock()
	defer s.RUnlock()

	has := true
	for _, item := range items {
		if _, has = s.m[item]; !has {
			break
		}
	}
	return has
}

// Size returns the number of items in a set.
func (s *setm[T]) Size() int {
	s.RLock()
	defer s.RUnlock()

	l := len(s.m)
	return l
}

// Clear removes all items from the set.
func (s *setm[T]) Clear() {
	s.Lock()
	defer s.Unlock()

	s.m = make(map[T]struct{})
}

// IsEqual test whether s and t are the same in size and have the same items.
func (s *setm[T]) IsEqual(t Set[T]) bool {
	s.RLock()
	defer s.RUnlock()

	// Force locking only if given set is threadsafe.
	if conv, ok := t.(rwLocker); ok {
		conv.RLock()
		defer conv.RUnlock()
	}

	// return false if they are no the same size
	if sameSize := len(s.m) == t.Size(); !sameSize {
		return false
	}

	equal := true
	t.Each(func(item T) bool {
		_, equal = s.m[item]
		return equal // if false, Each() will end
	})

	return equal
}

// IsSubset tests whether t is a subset of s.
func (s *setm[T]) IsSubset(t Set[T]) bool {
	s.RLock()
	defer s.RUnlock()

	return t.Each(func(item T) bool {
		_, ok := s.m[item]
		return ok
	})
}

func (s *setm[T]) Each(f func(item T) bool) bool {
	s.RLock()
	defer s.RUnlock()

	return s.set.Each(f)
}

// List returns a slice of all items.
func (s *setm[T]) List() []T {
	s.RLock()
	defer s.RUnlock()

	return maps.Keys(s.m)
}

func (s *setm[T]) Copy() Set[T] {
	u := newTS[T]()
	for item := range s.m {
		u.Add(item)
	}
	return u
}

func (s *setm[T]) Merge(t Set[T]) Set[T] {
	s.Lock()
	defer s.Unlock()

	t.Each(func(item T) bool {
		s.m[item] = null{}
		return true
	})

	return s
}
