package cmp

type Eq[T any] interface{ Eq(T) bool }

func Equal[T Eq[T]](a, b T) bool { return a.Eq(b) }

type Cmp[T any] interface{ Cmp(T) int }

func CompareType[T Cmp[T]](a, b T) int { return a.Cmp(b) }

func Equalizer[T any](c Cmp[T]) Eq[T] { return eq[T]{c} }

type eq[T any] struct{ Cmp[T] }

func (e eq[T]) Eq(t T) bool { return e.Cmp.Cmp(t) == 0 }
