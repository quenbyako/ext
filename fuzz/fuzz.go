package fuzz

import (
	"crypto/rand"
	"fmt"
	"io"
	"math"
	"math/big"

	"github.com/quenbyako/ext/slices"
)

type Fuzzer[T any] func(seed io.Reader) T

func Const[T any](value T) Fuzzer[T] { return func(io.Reader) T { return value } }

func Any[T any](f func(io.Reader) T) Fuzzer[any] {
	return func(seed io.Reader) any { return f(seed) }
}

func Uint32(min, max uint32) Fuzzer[uint32] {
	if min == max {
		return Const(min)
	}
	if min > max {
		panic(fmt.Sprintf("min > max: %v > %v", min, max))
	}

	return func(seed io.Reader) uint32 {
		l, err := rand.Int(seed, big.NewInt(int64(max-min)))
		if err != nil {
			panic(err)
		}
		return uint32(l.Uint64()) + min
	}
}

func Uint64(min, max uint64) Fuzzer[uint64] {
	if min == max {
		return Const(min)
	}
	if min > max {
		panic(fmt.Sprintf("min > max: %v > %v", min, max))
	}

	return func(seed io.Reader) uint64 {
		l, err := rand.Int(seed, big.NewInt(int64(max-min)))
		if err != nil {
			panic(err)
		}
		return l.Uint64() + min
	}
}

func Ptr[T any](chance float64, f Fuzzer[T]) Fuzzer[*T] {
	return func(seed io.Reader) *T {
		if Bool(chance)(seed) {
			return nil
		}

		return ptr(f(seed))
	}
}

// chance is 0 to 1
func Bool(chance float64) Fuzzer[bool] {
	return func(seed io.Reader) bool {
		return chance == 1 || Float64()(seed) < chance
	}
}

// any number from 0 to 1
func Float64() Fuzzer[float64] {
	return func(seed io.Reader) float64 {
		if v := float64(Uint64(0, math.MaxInt64)(seed)) / math.MaxInt64; !math.IsNaN(v) {
			return v
		}

		return 0
	}
}

// any number from 0 to 1
func Float32() Fuzzer[float32] {
	return func(seed io.Reader) float32 {
		if v := float32(Uint32(0, math.MaxInt32)(seed)) / math.MaxInt32; !math.IsNaN(float64(v)) {
			return v
		}

		return 0
	}
}

func Bytes(min, max uint64) Fuzzer[[]byte] {
	return func(seed io.Reader) []byte {
		l := Uint64(min, max)(seed)
		ret := make([]byte, l)
		seed.Read(ret)

		return ret
	}
}

func String(min, max uint64) Fuzzer[string] {
	return func(seed io.Reader) string {
		const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
		const len = len(letters)
		resultLen := Uint64(min, max)(seed)

		return string(slices.Generate(int(resultLen), func(int) byte {
			num, _ := rand.Int(seed, big.NewInt(int64(len)))
			return letters[num.Int64()]
		}))
	}
}

func Slice[T any](min, max int, f Fuzzer[T]) Fuzzer[[]T] {
	return func(seed io.Reader) []T {
		l := Uint64(uint64(min), uint64(max))(seed)
		s := make([]T, l)
		for i := range s {
			s[i] = f(seed)
		}

		return s
	}
}

func Map[K comparable, V any](min, max int, k Fuzzer[K], v Fuzzer[V]) Fuzzer[map[K]V] {
	return func(seed io.Reader) map[K]V {
		l := Uint64(uint64(min), uint64(max))(seed)
		m := make(map[K]V, l)
		for i := min; i < min+int(l); i++ {
			m[k(seed)] = v(seed)
		}

		return m
	}
}

func ptr[T any](v T) *T { return &v }
