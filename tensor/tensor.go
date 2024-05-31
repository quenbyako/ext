package tensor

import (
	"fmt"

	"github.com/quenbyako/ext/slices"
)

type Tensor[T any] interface {
	At(i ...int) T
	Shape() []int
}

type TensorWritable[T any] interface {
	Tensor[T]
	Set(t T, i ...int)
}

func String[T any](t Tensor[T]) string {
	shape := t.Shape()
	maxLen := 0
	strs := slices.Remap(Items(t), func(t T) string {
		str := fmt.Sprintf("%#v", t)
		maxLen = max(maxLen, len(str))
		return str
	})

	for i, item := range strs {
		strs[i] = fmt.Sprintf("%#v", item)

	}

	return strTensor(nil, shape, func(i ...int) string { return strs[index(shape, i)] }, maxLen)
}

func Size[T any](t Tensor[T]) int { shape := t.Shape(); return mul(shape[0], shape[1:]...) }

func Items[T any](t Tensor[T]) []T {
	shape := t.Shape()
	s := make([]T, mul(shape[0], shape[1:]...))
	for i := range len(s) {
		s[i] = t.At(reverseIndex(shape, i)...)
	}

	return s
}

func mul[T Number](x T, y ...T) T {
	if len(y) == 0 {
		return x
	}
	for _, y := range y {
		x *= y
	}
	return x
}

func sum[T Number](y ...T) (t T) {
	if len(y) == 0 {
		return t
	}

	for _, y := range y {
		t += y
	}

	return t
}

func sub[T Number](y ...T) (t T) {
	if len(y) == 0 {
		return t
	}

	for _, y := range y {
		t -= y
	}

	return t
}
