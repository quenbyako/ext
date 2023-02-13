package set

type Hashable interface {
	Hash() (uint64, error)
}

func mushHash(item Hashable) uint64 {
	h, err := item.Hash()
	if err != nil {
		panic(err)
	}
	return h
}

type setAny[T Hashable] map[uint64]T

func newAnyNonTS[T Hashable](items ...T) Set[T] { return make(setAny[T]).Add(items...) }

// Add includes the specified items (one or more) to the set. The underlying
// Set s is modified. If passed nothing it silently returns.
func (s setAny[T]) Add(items ...T) Set[T] {
	for _, item := range items {
		h, err := item.Hash()
		if err != nil {
			panic(err)
		}
		s[h] = item
	}

	return s
}

// Remove deletes the specified items from the set.  The underlying Set s is
// modified. If passed nothing it silently returns.
func (s setAny[T]) Remove(items ...T) Set[T] {
	for _, item := range items {
		delete(s, mushHash(item))
	}
	return s
}

// Pop  deletes and return an item from the set. The underlying Set s is
// modified. If set is empty, nil is returned.
func (s setAny[T]) Pop() (T, bool) {
	for h, item := range s {
		defer delete(s, h)
		return item, true
	}

	var t T

	return t, false
}

// Has looks for the existence of items passed. It returns false if nothing is
// passed. For multiple items it returns true only if all of  the items exist.
func (s setAny[T]) Has(items ...T) bool {
	// assume checked for empty item, which not exist
	if len(items) == 0 {
		return false
	}

	for _, item := range items {
		if _, ok := s[mushHash(item)]; !ok {
			return false
		}
	}
	return true
}

func (s setAny[T]) Size() int     { return len(s) }
func (s setAny[T]) Clear()        { s = make(map[uint64]T) }
func (s setAny[T]) IsEmpty() bool { return s.Size() == 0 }
func (s setAny[T]) IsEqual(t Set[T]) bool {
	// Force locking only if given set is threadsafe.
	if conv, ok := t.(rwLocker); ok {
		conv.RLock()
		defer conv.RUnlock()
	}

	// return false if they are no the same size
	if sameSize := len(s) == t.Size(); !sameSize {
		return false
	}

	return t.Each(func(item T) bool {
		_, ok := s[mushHash(item)]
		return ok // if false, Each() will end
	})
}

// IsSubset tests whether t is a subset of s.
func (s setAny[T]) IsSubset(t Set[T]) bool {
	return t.Each(func(item T) bool {
		_, ok := s[mushHash(item)]
		return ok
	})
}

// IsSuperset tests whether t is a superset of s.
func (s setAny[T]) IsSuperset(t Set[T]) bool { return t.IsSubset(s) }

// Each traverses the items in the Set, calling the provided function for each
// set member. Traversal will continue until all items in the Set have been
// visited, or if the closure returns false.
func (s setAny[T]) Each(f func(item T) bool) bool {
	for _, item := range s {
		if !f(item) {
			return false
		}
	}

	return true
}

// Copy returns a new Set with a copy of s.
func (s setAny[T]) Copy() Set[T] {
	u := make(setAny[T])
	for h, item := range s {
		u[h] = item
	}
	return u
}

// String returns a string representation of s
func (s setAny[T]) String() string { return stringSet[T](s) }

// List returns a slice of all items. There is also StringSlice() and
// IntSlice() methods for returning slices of type string or int.
func (s setAny[T]) List() []T {
	list := make([]T, 0, len(s))

	for _, item := range s {
		list = append(list, item)
	}

	return list
}

// Merge is like Union, however it modifies the current set it's applied on
// with the given t set.
func (s setAny[T]) Merge(t Set[T]) Set[T] {
	t.Each(func(item T) bool {
		s[mushHash(item)] = item
		return true
	})

	return s
}

// it's not the opposite of Merge.
// Separate removes the set items containing in t from set s. Please aware that
func (s setAny[T]) Separate(t Set[T]) Set[T] { return s.Remove(t.List()...) }
