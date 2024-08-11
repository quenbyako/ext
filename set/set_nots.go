package set

import "fmt"

// Provides a common set baseline for both threadsafe and non-ts Sets.
type set[T comparable] map[T]struct{} // struct{} doesn't take up space

var _ Set[int] = set[int](nil)

// NewNonTS creates and initializes a new non-threadsafe Set.
func newNonTS[T comparable](items ...T) Set[T] { return (make(set[T], len(items))).Add(items...) }

// Add includes the specified items (one or more) to the set. The underlying
// Set s is modified. If passed nothing it silently returns.
func (s set[T]) Add(items ...T) Set[T] {
	for _, item := range items {
		s[item] = null{}
	}

	return s
}

// Remove deletes the specified items from the set.  The underlying Set s is
// modified. If passed nothing it silently returns.
func (s set[T]) Del(items ...T) Set[T] {
	for _, item := range items {
		delete(s, item)
	}

	return s
}

// Has looks for the existence of items passed. It returns false if nothing is
// passed. For multiple items it returns true only if all of  the items exist.
func (s set[T]) Has(item T) bool {
	_, ok := s[item]

	return ok
}

func (s set[T]) Len() int { return len(s) }

func (s set[T]) Each(yield func(T) bool) {
	for item := range s {
		if !yield(item) {
			break
		}
	}
}

func (s set[T]) Format(f fmt.State, verb rune) {
	fmtValue := fmt.FormatString(f, verb)

	fmt.Fprintf(f, "set[")

	var (
		i    int //nolint:varnamelen
		last = s.Len() - 1
	)

	s.Each(func(b T) bool {
		fmt.Fprintf(f, fmtValue, b)

		if i++; i <= last {
			fmt.Fprintf(f, ", ")
		}

		return true
	})
	fmt.Fprintf(f, "]")
}
