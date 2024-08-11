// Package set provides both threadsafe and non-threadsafe implementations of
// a generic set data structure. In the threadsafe set, safety encompasses all
// operations on one set. Operations on multiple sets are consistent in that
// the elements of each set used was valid at exactly one point in time
// between the start and the end of the operation.
package set

// Set is describing a Set. Sets are an unordered, unique list of values.
type Set[T any] interface {
	Add(items ...T) Set[T]
	Del(items ...T) Set[T]

	// Has looks for the existence of items passed. It returns false if nothing
	// is passed. For multiple items it returns true only if all of  the items
	// exist.
	Has(items T) bool
	Len() int

	// Each traverses the items in the Set, calling the provided function for each
	// set member. Traversal will continue until all items in the Set have been
	// visited, or if the closure returns false.
	Each(yield func(T) bool)
}

type null = struct{}

// New creates and initializes a new Set interface. Its single parameter
// denotes the type of set to create. Either ThreadSafe or
// NonThreadSafe. The default is ThreadSafe.
func New[T comparable](items ...T) Set[T] { return newNonTS(items...) }

// NewTS creates and initializes a new Set with thread safety. The type must
// implement the [Hashable] interface.
//
// The behavior of thread safe version is equal to [sync.Map].
func NewTS[T Hashable](items ...T) Set[T] { return newTS(items...) }

// NewAny creates and initializes a new Set with any type of values. The
// type must implement the [Hashable] interface.
func NewAny[T Hashable](items ...T) Set[T] { return newAnyNonTS(items...) }

// Pop  deletes and return an item from the set. The underlying Set s is
// modified. If set is empty, nil is returned.
//
//nolint:ireturn,nonamedreturns // generic is not means that it's an interface
func Pop[T any](s Set[T]) (t T, ok bool) {
	s.Each(func(item T) bool {
		s.Del(t)
		t, ok = item, true

		return false
	})

	return t, ok
}

// Clone returns a new Set with the same items. The underlying Set s is not
// modified.
//
//nolint:varnamelen // logically reasonable
func Clone[T comparable](s Set[T]) Set[T] {
	c := make(set[T], s.Len())
	s.Each(func(t T) bool {
		c.Add(t)

		return true
	})

	return c
}

// IsEqual tests whether a and b are equal by containing same items.
//
//nolint:varnamelen
func IsEqual[T any](a, b Set[T]) bool {
	if a.Len() != b.Len() {
		return false
	}

	ok := true

	a.Each(func(item T) bool {
		if !b.Has(item) {
			ok = false
		}

		return ok
	})

	return ok
}

// IsSubset tests whether b is a subset of a.
//
//nolint:varnamelen
func IsSubset[T any](a, b Set[T]) bool {
	if a.Len() < b.Len() {
		return false
	}

	ok := true

	b.Each(func(item T) bool {
		if !a.Has(item) {
			ok = false
		}

		return ok
	})

	return ok
}

// IsSuperset tests whether b is a superset of a.
func IsSuperset[T any](a, b Set[T]) bool { return IsSubset(b, a) }

// List returns a slice of all items. There is also StringSlice() and
// IntSlice() methods for returning slices of type string or int.
//

func AsList[T any](s Set[T]) []T {
	list := make([]T, 0, s.Len())

	s.Each(func(item T) bool {
		list = append(list, item)

		return true
	})

	return list
}

// Union adds all items from t to s. The underlying Set a is modified.
//
//nolint:varnamelen // logically reasonable
func Union[T any](a Set[T], b ...Set[T]) Set[T] {
	for _, b := range b {
		b.Each(func(item T) bool {
			a.Add(item)

			return true
		})
	}

	return a
}

// Intersection returns a new Set with items that are in both a and b.
//
//nolint:varnamelen // logically reasonable
func Intersection[T any](a Set[T], b ...Set[T]) Set[T] {
	for _, b := range b {
		a.Each(func(item T) bool {
			if !b.Has(item) {
				a.Del(item)
			}

			return true
		})
	}

	return a
}

// Difference returns a new Set with items that are in a but not in b.
//
//nolint:varnamelen // logically reasonable
func Difference[T any](a Set[T], b ...Set[T]) Set[T] {
	for _, b := range b {
		b.Each(func(item T) bool {
			a.Del(item)

			return true
		})
	}

	return a
}

// SymmetricDifference returns a new Set with items that are in either a or b,
// but not in both.
//
//nolint:varnamelen
func SymmetricDifference[T any](a, b Set[T]) Set[T] {
	b.Each(func(item T) bool {
		if a.Has(item) {
			a.Del(item)
		} else {
			a.Add(item)
		}

		return true
	})

	return a
}
