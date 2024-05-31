package tensor

import (
	"fmt"
	"math"

	"github.com/quenbyako/ext/slices"
)

func main() { main1() }

func main2() {
	const (
		inputN  = 3
		hidden1 = 3
		hidden2 = 3

		learningRate = .01
	)

	// activation := func(n float64, i ...int) float64 { return n }
	activation := func(n float64, i ...int) float64 { return sigmoid(n) }

	x1 := NewFrom([]float64{.03, .72, .49}, 1, inputN)
	w1 := NewFrom([]float64{
		.88, .38, .90,
		.37, .14, .41,
		.96, .50, .60,
	}, inputN, hidden1)
	b1 := NewFrom([]float64{.23, .89, .08}, 1, hidden1)

	t1 := Add(Mul(w1, x1), b1)
	h1 := Remap(t1, activation)

	fmt.Println(String(h1))
	fmt.Println("_______________")

	w2 := NewFrom([]float64{
		.29, .57, .36,
		.73, .53, .68,
		.01, .02, .58,
	}, hidden1, hidden2)
	b2 := NewFrom([]float64{.78, .83, .8}, 1, hidden2)
	t2 := Add(Mul(w2, h1), b2)
	h2 := Remap(t2, activation)

	fmt.Println(String(h2))
	fmt.Println("_______________")

	want := NewFrom([]float64{.93, .74, .17}, 1, hidden2)

	delta := CalculateMSE(h2, want, meanError)

	deltaLong := StridedFunc(func(n ...int) float64 {
		return delta.At(0, n[0])
	}, hidden2, hidden2)

	fmt.Println(w2)
	fmt.Println(deltaLong)

	fmt.Println(Mul(w2, deltaLong))

	// fmt.Println(err.String())
}

func main1() {
	const (
		inputN  = 2
		hidden1 = 2
		hidden2 = 1

		learnRate = 0.05
	)

	activation := func(n float64, i ...int) float64 { return n }
	// activation := func(n float64, i ...int) float64 { return sigmoid(n) }

	x1 := NewFrom([]float64{2, 3}, 1, inputN)

	w1 := NewFrom([]float64{.11, .21, .12, .08}, inputN, hidden1)
	b1 := NewFrom([]float64{0, 0}, 1, hidden1)
	t1 := Add(Mul(w1, x1), b1)
	h1 := Remap(t1, activation)

	fmt.Println(String(h1))
	fmt.Println("_______________")

	w2 := NewFrom([]float64{.14, .15}, hidden1, hidden2)
	b2 := NewFrom([]float64{0}, 1, hidden2)
	t2 := Add(Mul(w2, h1), b2)
	h2 := Remap(t2, activation)

	fmt.Println(String(h2))
	fmt.Println("_______________")

	y := NewFrom([]float64{1}, 1, hidden2)

	// delta2 := CalculateMSE(h2, y, meanError)
	delta2 := Remap(Sub(h2, y), func(t float64, n ...int) float64 { return t * learnRate })
	fmt.Println(String(delta2))
	fmt.Println("______________")

	delta2Long := NewFunc([]int{hidden2, hidden2}, func(n ...int) float64 {
		return delta2.At(0, n[0])
	})

	fmt.Println("______________")
	fmt.Println(h1)
	fmt.Println(delta2Long)
	e1 := Transpose(Mul(h1, delta2Long))
	fmt.Println("Reduced:\n" + String(Sub(w2, e1)))
	fmt.Println("______________")
}

func sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

func sigmoidPrime(x float64) float64 {
	return x * (1.0 - x)
}

func meanError[T Number](pred, act T) T { return (pred - act) * (pred - act) / 2 }

func CalculateMSE[T Number](pred, act Tensor[T], f func(a, b T) T) Tensor[T] {
	if !slices.Equal(pred.Shape(), act.Shape()) {
		panic(fmt.Sprintf("tensors are not equal: a%v b%v", pred.Shape(), act.Shape()))
	}

	c := Zeros[T](pred.Shape()...)
	for i := range pred.values {
		c.values[i] = f(pred.values[i], act.values[i])
	}

	return c
}

// ┌─────┬────────┐
// | age | weight |
// ├─────┼────────┤
// │ 10  │   100  │
// │ 80  │  500   │
// └─────┴────────┘
