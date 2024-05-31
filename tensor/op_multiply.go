package tensor

import (
	"fmt"
	"strings"
)

func Product[T Number](a, b Tensor[T]) Tensor[T] {
	ashape := a.Shape()
	bshape := b.Shape()

	if len(ashape) != 2 {
		panic(fmt.Sprintf("multiplication of tensors with 3+ dimensions is not supported: got %v", len(ashape)))
	}
	if len(ashape) != len(bshape) || ashape[0] != bshape[1] {
		panic(fmt.Sprintf("invalid shape: a(%v), b(%v)", ashape, bshape))
	}

	shape := []int{bshape[0], ashape[1]}
	c := Zeros[T](shape...)
	for i1 := range shape[1] {
		for i0 := range shape[0] {
			var total T
			parts := make([]string, ashape[0])
			for j := range ashape[0] {
				parts[j] = fmt.Sprintf("%v*%v", a.At(j, i1), b.At(i0, j))

				total += a.At(j, i1) * b.At(i0, j)
			}

			fmt.Println(i0, i1, strings.Join(parts, " + "), "=", total)

			c.Set(total, i0, i1)
		}
	}

	return c
}
