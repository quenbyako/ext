//nolint:testpackage // utils for tests
package lru

import (
	"errors"
	"slices"
	"sync/atomic"
	"testing"
	"time"
)

func NewWithClock[K comparable, V any](
	constructor ConstructorFunc[K, V],
	validate ValidateFunc[V],
	onEvict EvictCallback[K, V],
	capacity int,
	ttl time.Duration,
) (cache *Cache[K, V], mClock *MockClock) {
	if validate == nil {
		validate = func(V) bool { return true }
	}

	mClock = &MockClock{
		now:           time.Now(),
		TickerCreated: make(chan struct{}),
	}

	return newCache(
		constructor,
		onEvict,
		validate,
		capacity,
		mClock,
		ttl,
	), mClock
}

func RequireNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func RequireErrorIs(t *testing.T, err, target error) {
	t.Helper()

	if !errors.Is(err, target) {
		t.Fatalf("unexpected error: %v, target: %v", err, target)
	}
}

func RequireEqual[T comparable](t *testing.T, actual, expected T) {
	t.Helper()

	if actual != expected {
		t.Fatalf("unexpected value: %v, expected: %v", actual, expected)
	}
}

func RequireError(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func RequireContains[S ~[]T, T comparable](t *testing.T, actual S, expected T) {
	t.Helper()

	for _, v := range actual {
		if v == expected {
			return
		}
	}

	t.Fatalf("expected slice to contain %v, got %v", expected, actual)
}

func RequireEqualSlice[S ~[]T, T comparable](t *testing.T, actual, expected S) {
	t.Helper()

	if !slices.Equal(actual, expected) {
		t.Fatalf("unexpected slice: %v, expected: %v", actual, expected)
	}
}

type MockClock struct {
	now           time.Time
	ch            chan time.Time
	tickerStarted atomic.Bool
	interval      time.Duration
	nextTick      time.Time
	TickerCreated chan struct{}
}

func (m *MockClock) Now() time.Time {
	return m.now
}

func (m *MockClock) TickerStarted() bool {
	return m.tickerStarted.Load()
}

func (m *MockClock) Ticker(dur time.Duration) (ch <-chan time.Time, stop func()) {
	if !m.tickerStarted.CompareAndSwap(false, true) {
		panic("ticker already started")
	}

	m.interval = dur
	m.nextTick = m.now.Add(dur)
	m.ch = make(chan time.Time, 1)

	if m.TickerCreated != nil {
		close(m.TickerCreated)
	}

	return m.ch, func() {}
}

func (m *MockClock) Advance(d time.Duration) {
	m.now = m.now.Add(d)

	if m.ch != nil && !m.now.Before(m.nextTick) {
		select {
		case m.ch <- m.nextTick:
		default:
		}

		m.nextTick = m.nextTick.Add(m.interval)
	}
}

func (m *MockClock) Close() {
	if m.ch != nil {
		close(m.ch)
		m.ch = nil
	}
}
