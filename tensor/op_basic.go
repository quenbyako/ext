package tensor

import (
	"fmt"

	"github.com/quenbyako/ext/slices"
)

func ItemsCount[T any](a Tensor[T]) int { s := a.Shape(); return mul(s[0], s[1:]...) }

func performN[T any](t []Tensor[T], f func(items ...T) T) Tensor[T] {
	s1 := t[0].Shape()
	sn := slices.Remap(t[1:], func(t Tensor[T]) []int { return t.Shape() })
	if !slicesEqualAll(s1, sn...) {
		panic(fmt.Sprintf("tensors are not equal: v", append([][]int{s1}, sn...)))
	}

	c := Zeros[T](s1...)
	for i := range ItemsCount(c) {
		i := reverseIndex(s1, i)
		values := slices.Remap(t, func(t Tensor[T]) T { return t.At(i...) })

		c.Set(f(values...), i...)
	}

	return c
}

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 |
		~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64
}

func perform[T Number](a, b Tensor[T], f func(a, b T) T) Tensor[T] {
	as, bs := a.Shape(), b.Shape()
	if !slices.Equal(as, bs) {
		panic(fmt.Sprintf("tensors are not equal: a%v b%v", as, bs))
	}

	c := Zeros[T](as...)
	for i := range ItemsCount(c) {
		i := reverseIndex(as, i)
		c.Set(f(a.At(i...), b.At(i...)))
	}

	return c
}

func performS[T Number](a Tensor[T], b T, f func(a, b T) T) Tensor[T] {
	shape := a.Shape()
	c := Zeros[T](shape...)
	for i := range ItemsCount(c) {
		i := reverseIndex(shape, i)
		c.Set(f(a.At(i...), b))
	}

	return c
}

func addOne[T Number](a, b T) T { return a + b }
func subOne[T Number](a, b T) T { return a - b }
func mulOne[T Number](a, b T) T { return a * b }
func divOne[T Number](a, b T) T { return a / b }

func AddScalar[T Number](a Tensor[T], b T) Tensor[T] { return performS(a, b, addOne) }
func SubScalar[T Number](a Tensor[T], b T) Tensor[T] { return performS(a, b, subOne) }
func MulScalar[T Number](a Tensor[T], b T) Tensor[T] { return performS(a, b, mulOne) }
func DivScalar[T Number](a Tensor[T], b T) Tensor[T] { return performS(a, b, divOne) }

func Add[T Number](a, b Tensor[T]) Tensor[T] { return perform(a, b, addOne) }
func Sub[T Number](a, b Tensor[T]) Tensor[T] { return perform(a, b, subOne) }
func Mul[T Number](a, b Tensor[T]) Tensor[T] { return perform(a, b, mulOne) }
func Div[T Number](a, b Tensor[T]) Tensor[T] {
	// for safety reasons, it's implemented differently: we are adding to
	// panic info about index

	as, bs := a.Shape(), b.Shape()
	if !slices.Equal(as, bs) {
		panic(fmt.Sprintf("tensors are not equal: a%v b%v", as, bs))
	}

	c := Zeros[T](as...)
	for i := range ItemsCount(c) {
		i := reverseIndex(as, i)
		defer func() {
			if p := recover(); p != nil {
				panic(fmt.Sprintf("%v: %v", as, p))
			}
		}()

		c.Set(a.At(i...) / b.At(i...))
	}

	return c

}

func slicesEqualAll[S ~[]E, E comparable](s1 S, sn ...S) bool {
	return sameFunc(slices.Equal, s1, sn...)
}

func same[E comparable](e E, en ...E) bool {
	return sameFunc(func(a, b E) bool { return a == b }, e, en...)
}

func sameFunc[E any](f func(a, b E) bool, e E, en ...E) bool {
	for _, item := range en {
		if !f(e, item) {
			return false
		}
	}

	return true
}
