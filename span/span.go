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
	// IMPORTANT:  unlike add, remove is non-included. In mathematical terms,
	// add is [ n : m ], and removing is ( n : m )
	Remove(Bound[T]) Span[T]
	In(T) bool
	Merge(Span[T]) Span[T]
	// проверяет, что ВЕСЬ интервал полностью входит в изучаемый, и не выходит
	// за границы
	Contains(Span[T]) bool
	ContainsBound(Bound[T]) bool
	Bounds() []Bound[T]
}

func New[T constraints.Ordered](bounds ...Bound[T]) Span[T] { return NewWithMerger(nil, bounds...) }

// NewWithMerger is absolutely same constructor as New, but allows to add a
// funtion, which checks, can we merge near bounds into single one. For example,
// if you created integer span, you can merge [0:2][3:4] as [0:4], which will
// mean completely same span.
func NewWithMerger[T constraints.Ordered](isSimilar func(previousEnd, nextStart T) bool, bounds ...Bound[T]) Span[T] {
	if isSimilar == nil {
		isSimilar = func(previousEnd, nextStart T) bool { return false }
	}

	var s Span[T] = &span[T]{
		isSimilar: isSimilar,
		bounds:    make([]Bound[T], 0, len(bounds)),
	}
	for _, b := range bounds {
		s = s.Add(b)
	}

	return s
}

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
		panic("lo is higher than hi")
	}
	return bound[T]{lo: lo, loIncluded: true, hi: hi, hiIncluded: true}
}

func ToBasic[T any](s Span[T]) [][2]T {
	return slices.Remap(s.Bounds(), func(b Bound[T]) [2]T { return [2]T{b.Lo(), b.Hi()} })
}

func FromBasicOrdered[T constraints.Ordered](s [][2]T) Span[T] {
	return New(slices.Remap(s, func(b [2]T) Bound[T] { return NewBound(b[0], b[1]) })...)
}

type span[T constraints.Ordered] struct {
	isSimilar func(previousEnd, nextStart T) bool
	bounds    []Bound[T]
}

func (s *span[T]) Bounds() []Bound[T] { return s.bounds }

// In tests whether a set contains an element. It has time complexity O(log r)
// where r is the number of ranges of the set (since it does a binary search
// over the ranges). In the worst case it is O(log n).
func (s *span[T]) In(b T) bool {
	_, ok := s.search(b)
	return ok
}

func (s *span[T]) Contains(y Span[T]) bool {
	return slices.IndexFunc(s.bounds, func(t Bound[T]) bool { return !s.ContainsBound(t) }) >= 0
}

func (s *span[T]) ContainsBound(y Bound[T]) bool {
	i, ok := s.search(y.Lo())

	return ok && s.bounds[i].Contains(y)
}

func (s *span[T]) Merge(y Span[T]) (z Span[T]) {
	z = s
	for _, b := range y.Bounds() {
		z = z.Add(b)
	}

	return z
}

func (s *span[T]) Add(y Bound[T]) Span[T] {
	if len(s.bounds) == 0 {
		return &span[T]{
			isSimilar: s.isSimilar,
			bounds:    []Bound[T]{y},
		}
	}

	// checking, that bound edges are already inside a span. Index is an index
	// of found bound
	loIndex, loOk := slices.BinarySearchFunc(s.bounds, nil, func(a, _ Bound[T]) int { return -a.Position(y.Lo()) })
	hiIndex, hiOk := slices.BinarySearchFunc(s.bounds, nil, func(a, _ Bound[T]) int { return -a.Position(y.Hi()) })

	if loOk && hiOk && loIndex == hiIndex {
		return s.copy()
	}

	// if lower bound edge has corresponding value to new bound, we will merge
	// them. Like [0:5] and [6:9] can be merged to [0:9]
	if !loOk && loIndex > 0 && s.isSimilar(s.bounds[loIndex-1].Hi(), y.Lo()) {
		loOk = true
		loIndex--
	}

	// same like above, but for higher numbers
	if !hiOk && hiIndex < len(s.bounds)-1 && s.isSimilar(y.Hi(), s.bounds[hiIndex+1].Lo()) {
		hiOk = true
	}

	z := &span[T]{
		isSimilar: s.isSimilar,
		bounds:    make([]Bound[T], loIndex),
	}
	copy(z.bounds, s.bounds[:loIndex])

	b := bound[T]{lo: y.Lo(), hi: y.Hi()}
	if loOk {
		b.lo = s.bounds[loIndex].Lo()
	}
	if hiOk {
		b.hi = s.bounds[hiIndex].Hi()
	}

	z.bounds = append(z.bounds, b)

	if b.hi < s.bounds[len(s.bounds)-1].Lo() {
		z.bounds = append(z.bounds, s.bounds[hiIndex:]...)
	}

	return z
}

func (s *span[T]) Remove(y Bound[T]) Span[T] {
	if len(s.bounds) == 0 {
		return &span[T]{
			isSimilar: s.isSimilar,
			bounds:    []Bound[T]{},
		}
	}

	if y.Lo() == y.Hi() || s.isSimilar(y.Lo(), y.Hi()) {
		panic(fmt.Sprintf("bound to remove must be open (n:m) instead of closed [n:m], which means that this bound is invalid: %v", y))
	}

	// checking, that bound edges are already inside a span. Index is an index
	// of found bound
	loIndex, loPos := s.getIndex(y.Lo())
	hiIndex, hiPos := s.getIndex(y.Hi())

	// В каких ситуациях баунд может существовать: (для удаления показано <n:m>)
	// * 1) баунд полностью сверху/снизу
	// * 2) баунд полностью перекрывает весь спан
	// * 3) баунд полностью между четким диапазоном (например '[0:1] >< [6:7]' '<3:4>' -> [0:1][6:7])
	// * 4) баунд частично сверху/снизу, кусочек накладывается на крайний баунд
	// * 5) баунд полностью внутри одного диапазона (например '> [0:10] <' '<3:4>' -> [0:3][4:10])
	// * 6) баунд между несколькими диапазонами  (например '[0:2][4:4][7:8]' '<3:4>' -> [0:3][4:10])
	switch {
	// case 1
	case hiPos == indexPositionLow || loPos == indexPositionHigh,
		// case 3
		hiPos == indexPositionBetween && loPos == indexPositionBetween && loIndex == hiIndex:
		return s

	// case 2
	case loPos == indexPositionLow && hiPos == indexPositionHigh:
		return NewWithMerger[T](s.isSimilar) // empty

	// case 4
	case loPos == indexPositionLow:
		var modified []Bound[T]
		if hiPos == indexPositionExact {
			modified = []Bound[T]{NewBound(y.Hi(), s.bounds[hiIndex].Hi())}
			hiIndex++
		}
		s.bounds = append(modified, s.bounds[hiIndex:]...)
		return s

	// case 4
	case hiPos == indexPositionHigh:
		var modified []Bound[T]
		if loPos == indexPositionExact {
			modified = []Bound[T]{NewBound(s.bounds[loIndex].Lo(), y.Lo())}
			loIndex--
		}
		s.bounds = append(s.bounds[:loIndex], modified...)
		return s

	// case 5
	case loPos == indexPositionExact && hiPos == indexPositionExact && loIndex == hiIndex:
		b := s.bounds[loIndex]
		s.bounds = slices.Replace(s.bounds, loIndex, hiIndex+1, NewBound(b.Lo(), y.Lo()), NewBound(y.Hi(), b.Hi()))
		return s

	// case 6
	default:
		// here we can be sure, that both of bounds to remove are inside span edges
		var modified []Bound[T]
		if loPos == indexPositionExact {
			modified = append(modified, NewBound(s.bounds[loIndex].Lo(), y.Lo()))
		}
		if hiPos == indexPositionExact {
			modified = append(modified, NewBound(y.Hi(), s.bounds[hiIndex].Hi()))
			hiIndex++
		}

		s.bounds = slices.Replace(s.bounds, loIndex, hiIndex, modified...)
		return s
	}
}

func (s *span[T]) copy() Span[T] {
	bounds := make([]Bound[T], len(s.bounds))
	copy(bounds, s.bounds)
	return &span[T]{isSimilar: s.isSimilar, bounds: bounds}
}

func (s *span[T]) edges() (lo, hi T) {
	if len(s.bounds) == 0 {
		panic("bounds are empty")
	}

	return s.bounds[0].Lo(), s.bounds[len(s.bounds)-1].Hi()
}

func (s *span[T]) search(value T) (int, bool) {
	return slices.BinarySearchFunc(s.bounds, value, func(a Bound[T], b T) int { return -a.Position(b) })
}

// если значение не в баундах, то возвращается ВЕРХНИЙ иднекс.
func (s *span[T]) getIndex(value T) (int, spanIndexPosition) {
	if len(s.bounds) == 0 {
		panic("bounds are empty")
	} else if s.bounds[0].Lo() > value {
		return 0, indexPositionLow
	} else if value > s.bounds[len(s.bounds)-1].Hi() {
		return len(s.bounds), indexPositionHigh
	} else if v, ok := s.search(value); ok {
		return v, indexPositionExact
	} else {
		return v, indexPositionBetween
	}
}

type spanIndexPosition uint8

const (
	indexPositionExact spanIndexPosition = iota
	// если возвращается between, то отдается правый (верхний) индекс
	indexPositionBetween
	indexPositionHigh
	indexPositionLow
)

func (s *span[T]) String() string { return joinStringer(s.bounds, "") }

type bound[T constraints.Ordered] struct {
	loIncluded, hiIncluded bool
	lo, hi                 T
}

func (x bound[T]) Lo() T                    { return x.lo }
func (x bound[T]) LoBetter() (T, bool)      { return x.lo, x.loIncluded }
func (x bound[T]) Hi() T                    { return x.hi }
func (x bound[T]) HiBetter() (T, bool)      { return x.hi, x.hiIncluded }
func (x bound[T]) Contains(y Bound[T]) bool { return x.lo >= y.Lo() && y.Hi() <= x.hi }
func (x bound[T]) Overlaps(y Bound[T]) bool { return !(x.lo > y.Hi() || x.hi < y.Lo()) }
func (x bound[T]) In(i T) bool              { return x.lo <= i && i <= x.hi }

func (x bound[T]) Position(i T) int {
	if i >= x.lo {
		if i > x.hi {
			return 1
		}
		return 0
	}
	return -1
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

func IsEqual[T comparable](a, b Span[T]) bool {
	aBounds, bBounds := a.Bounds(), b.Bounds()
	if len(aBounds) != len(bBounds) {
		return false
	}

	for i := range aBounds {
		if !IsBoundEqual(aBounds[i], bBounds[i]) {
			return false
		}
	}

	return true
}

func IsBoundEqual[T comparable](a, b Bound[T]) bool { return a.Lo() == b.Lo() && a.Hi() == b.Hi() }

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
