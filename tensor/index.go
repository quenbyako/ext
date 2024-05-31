package tensor

import "fmt"

func index(dims []int, n []int) (index int) {
	if len(dims) != len(n) {
		panic(fmt.Sprintf("invalid index shape (shape is %v, index has %v coordinates)", len(dims), len(n)))
	}

	for i := range dims {
		if n[i] >= dims[i] {
			panic(fmt.Sprintf("dimension %v is out of range (%v >= %v)", i, n[i], dims[i]))
		}
		index += n[i] * mul(dims[0], dims[1:i]...)
	}

	return index
}

func reverseIndex(dims []int, index int) []int {
	n := make([]int, len(dims))
	for i := range dims {
		n[i] = index % dims[i]
		index /= dims[i]
	}
	if index != 0 {
		panic("index out of range")
	}

	return n
}
