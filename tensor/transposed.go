package tensor

import "github.com/quenbyako/ext/slices"

type transposed[T any] struct {
	t Tensor[T]
}

var _ Tensor[int] = (*transposed[int])(nil)

func (m *transposed[T]) At(i ...int) T { return m.t.At(slices.Reverse(i)...) }
func (m *transposed[T]) Shape() []int  { return slices.Reverse(m.t.Shape()) }

func Transpose[T any](t Tensor[T]) Tensor[T] { return &transposed[T]{t: t} }
