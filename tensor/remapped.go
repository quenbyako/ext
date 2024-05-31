package tensor

import "github.com/quenbyako/ext/slices"

type remapped[T any] struct {
	t Tensor[T]
	f func(T, ...int) T
}

var _ Tensor[int] = (*remapped[int])(nil)

func (m *remapped[T]) At(i ...int) T { return m.f(m.t.At(slices.Reverse(i)...), i...) }
func (m *remapped[T]) Shape() []int  { return m.t.Shape() }

func Remap[T any](t Tensor[T], f func(T, ...int) T) Tensor[T] { return &remapped[T]{t: t, f: f} }
