package tensor

type strided[T any] struct {
	shape  []int
	values []T
}

func New[T any](shape ...int) TensorWritable[T] {
	return &strided[T]{
		shape:  shape,
		values: make([]T, mul(shape[0], shape[1:]...)),
	}
}

func NewFunc[T any](shape []int, at func(i ...int) T) TensorWritable[T] {
	s := &strided[T]{
		shape:  shape,
		values: make([]T, mul(shape[0], shape[1:]...)),
	}
	for i := range s.values {
		s.values[i] = at(reverseIndex(shape, i)...)
	}

	return s
}

func Zeros[T any](shape ...int) TensorWritable[T]  { return New[T](shape...) }
func Strided[T any](t Tensor[T]) TensorWritable[T] { return NewFunc(t.Shape(), t.At) }

func NewFrom[T Number](values []T, shape ...int) Tensor[T] {
	if len(values) != mul(shape[0], shape[1:]...) {
		panic("invalid shape: can't put values here")
	}

	return StridedFunc(func(n ...int) T { return values[index(shape, n)] }, shape...)
}

const floatPrecision = 1000000

func Randn[T float32 | float64](shape ...int) Tensor[T] {
	return StridedFunc(func(n ...int) T { return GetRandFloat[T](0, 1, floatPrecision) }, shape...)
}

func Range[T Number](shape ...int) Tensor[T] {
	return StridedFunc(func(n ...int) T { return T(index(shape, n)) }, shape...)
}

func StridedFunc[T any](f func(n ...int) T, shape ...int) Tensor[T] {
	values := make([]T, mul(shape[0], shape[1:]...))
	for i := range values {
		values[i] = f(reverseIndex(shape, i)...)
	}

	return &strided[T]{
		shape:  shape,
		values: values,
	}
}

func (m *strided[T]) At(i ...int) T     { return m.values[index(m.shape, i)] }
func (m *strided[T]) Shape() []int      { return m.shape }
func (m *strided[T]) Set(t T, i ...int) { m.values[index(m.shape, i)] = t }
