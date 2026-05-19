package lru_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/quenbyako/ext/container/lru"
)

const (
	testVal = "val"
)

var errDummy = errors.New("dummy error")

func TestCache_Basic(t *testing.T) {
	t.Parallel()

	var evictCalls atomic.Int32

	cacheInstance := New(func(ctx context.Context, key int) (string, error) {
		return testVal, nil
	}, nil, func(key int, val string) {
		evictCalls.Add(1)
	}, 10, 0)

	t.Cleanup(cacheInstance.Close)

	ctx := t.Context()
	val, put, err := cacheInstance.Get(ctx, 1)
	RequireNoError(t, err)
	RequireEqual(t, testVal, val)
	put()

	val, put, err = cacheInstance.Get(ctx, 1)
	RequireNoError(t, err)
	RequireEqual(t, testVal, val)
	put()

	RequireEqual(t, int32(0), evictCalls.Load())
}

func TestCache_EvictionCapacity(t *testing.T) {
	t.Parallel()

	var evicted []int

	cacheInstance := New(func(ctx context.Context, key int) (int, error) {
		return key * 10, nil
	}, nil, func(key, val int) {
		evicted = append(evicted, key)
	}, 2, 0)

	t.Cleanup(cacheInstance.Close)

	ctx := t.Context()
	_, put1, err := cacheInstance.Get(ctx, 1)
	RequireNoError(t, err)
	put1()

	_, put2, err := cacheInstance.Get(ctx, 2)
	RequireNoError(t, err)
	put2()

	_, put3, err := cacheInstance.Get(ctx, 3)
	RequireNoError(t, err)
	put3()

	RequireEqualSlice(t, []int{1}, evicted)
}

func TestCache_EvictionTTL(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	cacheInstance, mClock := NewWithClock(
		func(ctx context.Context, key int) (string, error) {
			calls.Add(1)
			return testVal, nil
		},
		nil,
		nil,
		10,
		10*time.Millisecond,
	)

	t.Cleanup(func() {
		cacheInstance.Close()
		mClock.Close()
	})

	ctx := t.Context()
	_, put1, err := cacheInstance.Get(ctx, 1)
	RequireNoError(t, err)
	put1()

	mClock.Advance(20 * time.Millisecond)

	_, put2, err := cacheInstance.Get(ctx, 1)
	RequireNoError(t, err)
	put2()
	RequireEqual(t, int32(2), calls.Load())
}

func TestCache_ConstructorError(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32

	cacheInstance := New(func(ctx context.Context, key int) (string, error) {
		if callCount.Add(1) == 1 {
			return "", errDummy
		}

		return "ok", nil
	}, nil, nil, 10, 0)

	t.Cleanup(cacheInstance.Close)

	ctx := t.Context()
	_, _, err := cacheInstance.Get(ctx, 1)
	RequireErrorIs(t, err, errDummy)

	val, put, err := cacheInstance.Get(ctx, 1)
	RequireNoError(t, err)
	RequireEqual(t, "ok", val)
	put()
}

func TestCache_ContextCancellation(t *testing.T) {
	t.Parallel()

	started, block := make(chan struct{}), make(chan struct{})
	cacheInstance := New(func(ctx context.Context, key int) (string, error) {
		close(started)
		<-block

		return "ready", nil
	}, nil, nil, 10, 0)

	t.Cleanup(cacheInstance.Close)

	ctx, cancel := context.WithCancel(t.Context())
	go func() {
		_, _, err := cacheInstance.Get(ctx, 1)
		RequireError(t, err)
	}()

	<-started
	cancel()
	close(block)
}

func TestCache_Close(t *testing.T) {
	t.Parallel()

	var evictCalls atomic.Int32

	cacheInstance := New(func(ctx context.Context, key int) (string, error) {
		return testVal, nil
	}, nil, func(key int, val string) {
		evictCalls.Add(1)
	}, 10, 0)

	ctx := t.Context()
	_, put1, err := cacheInstance.Get(ctx, 1)
	RequireNoError(t, err)
	put1()

	_, put2, err := cacheInstance.Get(ctx, 2)
	RequireNoError(t, err)
	put2()

	cacheInstance.Close()
	RequireEqual(t, int32(2), evictCalls.Load())

	_, _, err = cacheInstance.Get(ctx, 1)
	RequireErrorIs(t, err, ErrClosed)
}

func TestCache_Validation(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	valid := true
	cacheInstance := New(func(ctx context.Context, key int) (string, error) {
		calls.Add(1)
		return testVal, nil
	}, func(val string) bool {
		return valid
	}, nil, 10, 0)

	t.Cleanup(cacheInstance.Close)

	ctx := t.Context()
	_, put1, err := cacheInstance.Get(ctx, 1)
	RequireNoError(t, err)
	put1()
	RequireEqual(t, int32(1), calls.Load())

	valid = false
	_, put2, err := cacheInstance.Get(ctx, 1)
	RequireNoError(t, err)
	put2()
	RequireEqual(t, int32(2), calls.Load())
}

func TestCache_FakeClock(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	cacheInstance, mClock := NewWithClock(
		func(ctx context.Context, key int) (string, error) {
			calls.Add(1)
			return testVal, nil
		},
		nil,
		nil,
		10,
		10*time.Second,
	)

	t.Cleanup(func() {
		cacheInstance.Close()
		mClock.Close()
	})

	ctx := t.Context()
	_, put1, err := cacheInstance.Get(ctx, 1)
	RequireNoError(t, err)
	put1()
	RequireEqual(t, int32(1), calls.Load())

	mClock.Advance(11 * time.Second)

	_, put2, err := cacheInstance.Get(ctx, 1)
	RequireNoError(t, err)
	put2()
	RequireEqual(t, int32(2), calls.Load())
}

func TestCache_Janitor(t *testing.T) {
	t.Parallel()

	evictedChan := make(chan int, 1)

	cacheInstance, mClock := NewWithClock(
		func(ctx context.Context, key int) (string, error) {
			return testVal, nil
		},
		nil,
		func(key int, val string) {
			evictedChan <- key
		},
		10,
		10*time.Second,
	)

	t.Cleanup(func() {
		cacheInstance.Close()
		mClock.Close()
	})

	// Wait for the janitor goroutine to start and register its ticker.
	<-mClock.TickerCreated

	ctx := t.Context()
	_, put, err := cacheInstance.Get(ctx, 1)
	RequireNoError(t, err)
	put()

	mClock.Advance(11 * time.Second)

	RequireEqual(t, <-evictedChan, 1)
}

func TestCache_PinningPreventsEviction(t *testing.T) {
	t.Parallel()

	var evicted []int

	cacheInstance := New(func(ctx context.Context, key int) (int, error) {
		return key * 10, nil
	}, nil, func(key, val int) {
		evicted = append(evicted, key)
	}, 2, 0)

	t.Cleanup(cacheInstance.Close)

	ctx := t.Context()
	_, put1, err := cacheInstance.Get(ctx, 1)
	RequireNoError(t, err)
	_, put2, err := cacheInstance.Get(ctx, 2)
	RequireNoError(t, err)
	put2()

	_, put3, err := cacheInstance.Get(ctx, 3)
	RequireNoError(t, err)
	put3()

	RequireEqualSlice(t, []int{2}, evicted)
	put1()
}

func TestCache_ErrFull(t *testing.T) {
	t.Parallel()

	cacheInstance := New(func(ctx context.Context, key int) (string, error) {
		return testVal, nil
	}, nil, nil, 2, 0)

	t.Cleanup(cacheInstance.Close)

	ctx := t.Context()
	_, put1, err := cacheInstance.Get(ctx, 1)
	RequireNoError(t, err)
	_, put2, err := cacheInstance.Get(ctx, 2)
	RequireNoError(t, err)

	_, _, err = cacheInstance.Get(ctx, 3)
	RequireErrorIs(t, err, ErrFull)

	put1()
	put2()
}

func TestCache_CancelWaiter_KeepsSuccessfulLoadedValue(t *testing.T) {
	t.Parallel()

	started := make(chan struct{})
	finish := make(chan struct{})
	evictedChan := make(chan int, 1)

	var calls atomic.Int32

	// We set capacity to 1 so that when item 1 is successfully restored to the cache,
	// it will evict item 2, which triggers the evictedChan notification.
	cache := New(func(ctx context.Context, key int) (string, error) {
		if key == 1 {
			calls.Add(1)
			close(started)
			<-finish

			return "successful-value", nil
		}

		return "value-2", nil
	}, nil, func(key int, val string) {
		if key == 2 {
			evictedChan <- key
		}
	}, 1, 0)
	t.Cleanup(cache.Close)
	ctx, cancel := context.WithCancel(t.Context())
	getErrChan := make(chan error, 1)

	// Step 1: Start Get(1), which blocks in the constructor.
	go func() {
		_, put, err := cache.Get(ctx, 1)
		if err == nil {
			put()
		}

		getErrChan <- err
	}()

	// Step 2: Wait for constructor to start, then cancel Get(1).
	<-started
	cancel()
	RequireError(t, <-getErrChan)

	// Step 3: Pre-populate the cache with key 2 AFTER key 1 has been cancelled and deleted.
	// Since the cache is empty now, this won't trigger any evictions yet.
	_, put2, err := cache.Get(t.Context(), 2)
	RequireNoError(t, err)
	put2() // Unpin it so it can be evicted.

	// Step 4: Allow constructor to finish, and wait for key 2 to be evicted.
	// This proves that key 1 has been successfully inserted into the cache.
	close(finish)
	RequireEqual(t, <-evictedChan, 2)

	// Step 5: Get(1) should now return the successfully loaded value from cache.
	val, put, err := cache.Get(t.Context(), 1)
	RequireNoError(t, err)
	put()

	RequireEqual(t, val, "successful-value")
	RequireEqual(t, calls.Load(), int32(1))
}
