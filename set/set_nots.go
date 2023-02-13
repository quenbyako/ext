package set

// Provides a common set baseline for both threadsafe and non-ts Sets.
type set[T comparable] struct {
	m map[T]struct{} // struct{} doesn't take up space
}

var _ Set[int] = (*set[int])(nil)

// NewNonTS creates and initializes a new non-threadsafe Set.
func newNonTS[T comparable](items ...T) Set[T] { return (&set[T]{make(map[T]struct{})}).Add(items...) }

// Add includes the specified items (one or more) to the set. The underlying
// Set s is modified. If passed nothing it silently returns.
func (s *set[T]) Add(items ...T) Set[T] {
	for _, item := range items {
		s.m[item] = null{}
	}

	return s
}

// Remove deletes the specified items from the set.  The underlying Set s is
// modified. If passed nothing it silently returns.
func (s *set[T]) Remove(items ...T) Set[T] {
	for _, item := range items {
		delete(s.m, item)
	}
	return s
}

// Pop  deletes and return an item from the set. The underlying Set s is
// modified. If set is empty, nil is returned.
func (s *set[T]) Pop() (T, bool) {
	for item := range s.m {
		delete(s.m, item)
		return item, true
	}

	var t T

	return t, false
}

// Has looks for the existence of items passed. It returns false if nothing is
// passed. For multiple items it returns true only if all of  the items exist.
func (s *set[T]) Has(items ...T) bool {
	// assume checked for empty item, which not exist
	if len(items) == 0 {
		return false
	}

	for _, item := range items {
		if _, ok := s.m[item]; !ok {
			return false
		}
	}
	return true
}

func (s *set[T]) Size() int     { return len(s.m) }
func (s *set[T]) Clear()        { s.m = make(map[T]struct{}) }
func (s *set[T]) IsEmpty() bool { return s.Size() == 0 }
func (s *set[T]) IsEqual(t Set[T]) bool {
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
func (s *set[T]) IsSubset(t Set[T]) bool {
	return t.Each(func(item T) bool {
		_, ok := s.m[item]
		return ok
	})
}

// IsSuperset tests whether t is a superset of s.
func (s *set[T]) IsSuperset(t Set[T]) bool { return t.IsSubset(s) }

// Each traverses the items in the Set, calling the provided function for each
// set member. Traversal will continue until all items in the Set have been
// visited, or if the closure returns false.
func (s *set[T]) Each(f func(item T) bool) bool {
	for item := range s.m {
		if !f(item) {
			return false
		}
	}

	return true
}

// Copy returns a new Set with a copy of s.
func (s *set[T]) Copy() Set[T] {
	u := newNonTS[T]()
	for item := range s.m {
		u.Add(item)
	}
	return u
}

// String returns a string representation of s
func (s *set[T]) String() string { return stringSet[T](s) }

// List returns a slice of all items. There is also StringSlice() and
// IntSlice() methods for returning slices of type string or int.
func (s *set[T]) List() []T {
	list := make([]T, 0, len(s.m))

	for item := range s.m {
		list = append(list, item)
	}

	return list
}

// Merge is like Union, however it modifies the current set it's applied on
// with the given t set.
func (s *set[T]) Merge(t Set[T]) Set[T] {
	t.Each(func(item T) bool {
		s.m[item] = null{}
		return true
	})

	return s
}

// it's not the opposite of Merge.
// Separate removes the set items containing in t from set s. Please aware that
func (s *set[T]) Separate(t Set[T]) Set[T] { return s.Remove(t.List()...) }
