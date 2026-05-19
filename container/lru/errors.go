package lru

import (
	"errors"
)

var (
	ErrClosed = errors.New("cache is closed")
	ErrFull   = errors.New("cache is full: all elements are pinned")
)
