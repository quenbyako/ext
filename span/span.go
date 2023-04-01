package span

import (
	"fmt"
	"strings"

	"github.com/quenbyako/ext/constraints"
	"github.com/quenbyako/ext/slices"
)

func Merge[T constraints.Ordered](s ...Span[T]) Span[T] {
	if len(s) == 0 {
		return New[T]()
	}

	res := s[0]
	if res == nil {
		res = New[T]()
	}

	for _, s := range s[1:] {
		if s == nil {
			continue
		}
		res = res.Merge(s)
	}

	return res
}

type Span[T any] interface {
	Add(Bound[T]) Span[T]
	In(T) bool
	Merge(Span[T]) Span[T]
	// проверяет, что ВЕСЬ интервал полностью входит в изучаемый, и не выходит
	// за границы
	Contains(Span[T]) bool
	ContainsBound(Bound[T]) bool
	Bounds() []Bound[T]
}

func New[T constraints.Ordered](bounds ...Bound[T]) Span[T] { return span[T](bounds) }

type Bound[T any] interface {
	Lo() T
	Hi() T

	// проверяет что y полностью входит в x
	Contains(y Bound[T]) bool
	// проверяет пересечение
	In(i T) bool
	Join(y Bound[T]) (z Bound[T], ok bool)
	Overlaps(y Bound[T]) bool
	Position(i T) int

	fmt.Stringer
}

func NewBound[T constraints.Ordered](lo, hi T) Bound[T] {
	if lo > hi {
		return nil
	}
	return bound[T]{lo: lo, hi: hi}
}

func ToBasic[T any](s Span[T]) [][2]T {
	return slices.Remap(s.Bounds(), func(b Bound[T]) [2]T { return [2]T{b.Lo(), b.Hi()} })
}

func FromBasicOrdered[T constraints.Ordered](s [][2]T) Span[T] {
	return New(slices.Remap(s, func(b [2]T) Bound[T] { return NewBound(b[0], b[1]) })...)
}

type span[T constraints.Ordered] []Bound[T]

func (s span[T]) Bounds() []Bound[T] { return s }

// In tests whether a set contains an element. It has time complexity O(log r)
// where r is the number of ranges of the set (since it does a binary search
// over the ranges). In the worst case it is O(log n).
func (s span[T]) In(b T) bool {
	_, ok := s.search(b)
	return ok
}

func (s span[T]) Contains(y Span[T]) bool {
	return slices.IndexFunc(s, func(t Bound[T]) bool { return !s.ContainsBound(t) }) >= 0
}

func (s span[T]) ContainsBound(y Bound[T]) bool {
	i, ok := s.search(y.Lo())

	return ok && s[i].Contains(y)
}

func (x span[T]) Merge(y Span[T]) (z Span[T]) {
	z = x
	for _, b := range y.Bounds() {
		z = z.Add(b)
	}

	return z
}

func (x span[T]) Add(y Bound[T]) Span[T] {
	if len(x) == 0 {
		return span[T]{y}
	}

	loIndex, loOk := slices.BinarySearchFunc(x, nil, func(a, _ Bound[T]) int { return -a.Position(y.Lo()) })
	hiIndex, hiOk := slices.BinarySearchFunc(x, nil, func(a, _ Bound[T]) int { return -a.Position(y.Hi()) })




	z := make(span[T], loIndex)
	copy(z, x[:loIndex])

	b := bound[T]{lo: y.Lo(), hi: y.Hi()}
	if loOk {
		b.lo = x[loIndex].Lo()
	}
	if hiOk {
		b.hi = x[hiIndex].Hi()
	}

	z = append(z, b)

	if b.hi < x[len(x)-1].Lo() {
		z = append(z, x[hiIndex:]...)
	}

	return z
}

func (s span[T]) search(b T) (int, bool) {
	return slices.BinarySearchFunc(s, nil, func(a, _ Bound[T]) int { return -a.Position(b) })
}

func (s span[T]) String() string { return joinStringer(s, "") }

type bound[T constraints.Ordered] struct{ lo, hi T }

func (x bound[T]) Lo() T                    { return x.lo }
func (x bound[T]) Hi() T                    { return x.hi }
func (x bound[T]) Contains(y Bound[T]) bool { return x.lo >= y.Lo() && y.Hi() <= x.hi }
func (x bound[T]) Overlaps(y Bound[T]) bool { return !(x.lo > y.Hi() || x.hi < y.Lo()) }
func (x bound[T]) In(i T) bool              { return x.lo <= i && i <= x.hi }

func (x bound[T]) Position(i T) int {
	if x.hi < i {
		return 1
	} else if x.lo > i {
		return -1
	} else {
		return 0
	}
}

// возвращает true, если интервалы пересекаются и создался новый интервал
func (x bound[T]) Join(y Bound[T]) (z Bound[T], ok bool) {
	if !x.Overlaps(y) {
		return bound[T]{}, false
	} else if x.Contains(y) {
		return x, true
	} else if y.Contains(x) {
		return y, true
	}

	// now it overlaps by part, so bound must be modified
	return bound[T]{lo: min(x.lo, y.Lo()), hi: max(x.hi, y.Hi())}, true
}

func (x bound[T]) String() string { return fmt.Sprintf("[%v:%v]", x.lo, x.hi) }

func min[T constraints.Ordered](i ...T) T {
	var t T
	for _, i := range i {
		if i < t {
			t = i
		}
	}
	return t
}

func max[T constraints.Ordered](i ...T) T {
	var t T
	for _, i := range i {
		if i > t {
			t = i
		}
	}
	return t
}

func joinStringer[S ~[]T, T fmt.Stringer](s S, sep string) string {
	return strings.Join(slices.Remap(s, func(s T) string { return s.String() }), sep)
}
