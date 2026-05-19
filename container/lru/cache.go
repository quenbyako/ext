// Package lru implements a thread-safe LRU cache with duplicate suppression.
package lru

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type (
	ConstructorFunc[K comparable, V any] func(context.Context, K) (V, error)
	EvictCallback[K comparable, V any]   func(key K, value V)
	ValidateFunc[V any]                  func(value V) bool
)

type element[K comparable, V any] struct {
	next        *element[K, V]
	prev        *element[K, V]
	cancel      context.CancelFunc
	ch          chan struct{}
	err         error
	key         K
	value       V
	expiresAt   time.Time
	waiterCount int
	pinCount    int
	loading     bool
	discarded   bool
}

type Cache[K comparable, V any] struct {
	constructor ConstructorFunc[K, V]
	onEvict     EvictCallback[K, V]
	validate    ValidateFunc[V]
	clock       clock
	items       map[K]*element[K, V]
	head        *element[K, V]
	tail        *element[K, V]
	stopJanitor chan struct{}
	ttl         time.Duration
	capacity    int
	listLen     int
	mu          sync.Mutex
	closed      bool
}

func New[K comparable, V any](
	constructor ConstructorFunc[K, V],
	validate ValidateFunc[V],
	evict EvictCallback[K, V],
	capacity int,
	ttl time.Duration,
) *Cache[K, V] {
	if constructor == nil {
		constructor = func(context.Context, K) (V, error) { return *new(V), nil }
	}

	if validate == nil {
		validate = func(V) bool { return true }
	}

	if evict == nil {
		evict = func(K, V) {}
	}

	return newCache(constructor, evict, validate, capacity, systemClock{}, ttl)
}

func newCache[K comparable, V any](
	constructor ConstructorFunc[K, V],
	onEvict EvictCallback[K, V],
	validate ValidateFunc[V],
	capacity int,
	clock clock,
	ttl time.Duration,
) *Cache[K, V] {
	cache := &Cache[K, V]{
		constructor: constructor,
		onEvict:     onEvict,
		validate:    validate,
		clock:       clock,
		items:       make(map[K]*element[K, V]),
		head:        nil,
		tail:        nil,
		stopJanitor: make(chan struct{}),
		ttl:         ttl,
		capacity:    capacity,
		listLen:     0,
		mu:          sync.Mutex{},
		closed:      false,
	}

	if cache.ttl > 0 {
		go cache.janitor(cache.ttl)
	}

	return cache
}

func (c *Cache[K, V]) Get(ctx context.Context, key K) (val V, put func(), err error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return *new(V), nil, ErrClosed
	}

	elem, ok := c.items[key]
	if !ok {
		return c.startLoading(ctx, key)
	}

	if elem.loading {
		return c.waitLoading(ctx, key, elem)
	}

	if (c.isExpired(elem) || !c.validate(elem.value)) && elem.pinCount == 0 {
		c.evictElement(key, elem)
		return c.startLoading(ctx, key)
	}

	elem.pinCount++
	c.moveToFront(elem)
	c.mu.Unlock()

	return elem.value, c.makePutFunc(key, elem), nil
}

func (c *Cache[K, V]) isExpired(elem *element[K, V]) bool {
	return !elem.expiresAt.IsZero() && c.clock.Now().After(elem.expiresAt)
}

func (c *Cache[K, V]) evictElement(key K, elem *element[K, V]) {
	c.removeElement(elem)

	if c.items != nil {
		delete(c.items, key)
	}

	if c.onEvict != nil {
		c.onEvict(key, elem.value)
	}
}

func (c *Cache[K, V]) waitLoading(
	ctx context.Context,
	key K,
	elem *element[K, V],
) (val V, put func(), err error) {
	elem.waiterCount++
	elem.pinCount++
	completionChan := elem.ch
	c.mu.Unlock()

	select {
	case <-ctx.Done():
		c.cancelWaiter(key, elem)

		return *new(V), nil, fmt.Errorf("wait aborted: %w", ctx.Err())
	case <-completionChan:
		return c.resolveReady(key, elem)
	}
}

func (c *Cache[K, V]) cancelWaiter(key K, elem *element[K, V]) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem.waiterCount--
	elem.pinCount--

	if elem.waiterCount == 0 {
		elem.cancel()

		if c.items[key] == elem {
			delete(c.items, key)
		}
	}
}

func (c *Cache[K, V]) resolveReady(key K, elem *element[K, V]) (val V, put func(), err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.items[key] == elem && !elem.loading {
		c.moveToFront(elem)

		return elem.value, c.makePutFunc(key, elem), nil
	}

	elem.pinCount--

	return elem.value, nil, elem.err
}

func (c *Cache[K, V]) findAndEvictOldestUnpinned() bool {
	curr := c.tail
	for curr != nil {
		if curr.pinCount == 0 {
			c.evictElement(curr.key, curr)

			return true
		}

		curr = curr.prev
	}

	return false
}

// note: DO NOT remove type receiver here. It's an optimization to avoid generic
// type instantiation on every call from runtime.
func (*Cache[K, V]) makeLoadingElement(cancel context.CancelFunc, key K) *element[K, V] {
	return &element[K, V]{
		next:        nil,
		prev:        nil,
		cancel:      cancel,
		ch:          make(chan struct{}),
		err:         nil,
		key:         key,
		value:       *new(V),
		expiresAt:   time.Time{},
		waiterCount: 1,
		pinCount:    1,
		loading:     true,
		discarded:   false,
	}
}

func (c *Cache[K, V]) startLoading(ctx context.Context, key K) (val V, put func(), err error) {
	if c.capacity > 0 && len(c.items) >= c.capacity && !c.findAndEvictOldestUnpinned() {
		c.mu.Unlock()
		return *new(V), nil, ErrFull
	}

	loaderCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	elem := c.makeLoadingElement(cancel, key)
	c.items[key] = elem
	c.mu.Unlock()

	go c.runConstructor(loaderCtx, key, elem)

	select {
	case <-ctx.Done():
		c.cancelWaiter(key, elem)
		return *new(V), nil, fmt.Errorf("loading aborted: %w", ctx.Err())
	case <-elem.ch:
		return c.resolveReady(key, elem)
	}
}

func (c *Cache[K, V]) runConstructor(ctx context.Context, key K, elem *element[K, V]) {
	val, err := c.constructor(ctx, key)

	c.mu.Lock()
	defer c.mu.Unlock()

	elem.cancel()

	if err != nil {
		c.handleConstructorError(key, elem, err)

		return
	}

	c.handleConstructorSuccess(key, elem, val)
}

func (c *Cache[K, V]) handleConstructorError(key K, elem *element[K, V], err error) {
	elem.err = err

	if c.items[key] == elem {
		delete(c.items, key)
	}

	close(elem.ch)
}

func (c *Cache[K, V]) handleConstructorSuccess(key K, elem *element[K, V], val V) {
	elem.value = val

	elem.loading = false
	if c.ttl > 0 {
		elem.expiresAt = c.clock.Now().Add(c.ttl)
	}

	if c.closed || elem.discarded {
		c.evictOnFailure(key, val, elem.ch)

		return
	}

	if c.items[key] != elem {
		if c.items[key] != nil {
			c.evictOnFailure(key, val, elem.ch)

			return
		}

		c.items[key] = elem
	}

	if c.capacity > 0 && c.listLen >= c.capacity {
		c.evictOldest()
	}

	c.pushFront(elem)
	close(elem.ch)
}

func (c *Cache[K, V]) evictOnFailure(key K, val V, ch chan struct{}) {
	if c.onEvict != nil {
		c.onEvict(key, val)
	}

	close(ch)
}

func (c *Cache[K, V]) makePutFunc(key K, elem *element[K, V]) func() {
	var once sync.Once

	return func() {
		once.Do(func() {
			c.mu.Lock()
			defer c.mu.Unlock()

			elem.pinCount--

			if elem.pinCount == 0 && !elem.loading {
				if c.closed || c.isExpired(elem) || !c.validate(elem.value) {
					c.evictElement(key, elem)
				}
			}
		})
	}
}

func (c *Cache[K, V]) Remove(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if !ok {
		return false
	}

	if elem.loading {
		elem.discarded = true
	} else {
		c.removeAndEvict(key, elem)
	}

	delete(c.items, key)

	return true
}

func (c *Cache[K, V]) removeAndEvict(key K, elem *element[K, V]) {
	c.removeElement(elem)

	if c.onEvict != nil {
		c.onEvict(key, elem.value)
	}
}

func (c *Cache[K, V]) Close() {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}

	c.closed = true
	close(c.stopJanitor)

	for key, elem := range c.items {
		c.closeElement(key, elem)
	}
	c.mu.Unlock()
}

func (c *Cache[K, V]) closeElement(key K, elem *element[K, V]) {
	if elem.loading {
		elem.cancel()
		elem.discarded = true

		return
	}

	if elem.pinCount == 0 {
		c.removeElement(elem)
		delete(c.items, key)

		if c.onEvict != nil {
			c.onEvict(key, elem.value)
		}
	}
}

func (c *Cache[K, V]) evictOldest() {
	if c.tail == nil {
		return
	}

	oldest := c.tail
	c.removeElement(oldest)
	delete(c.items, oldest.key)

	if c.onEvict != nil {
		c.onEvict(oldest.key, oldest.value)
	}
}

func (c *Cache[K, V]) removeElement(elem *element[K, V]) {
	if elem.prev != nil {
		elem.prev.next = elem.next
	} else {
		c.head = elem.next
	}

	if elem.next != nil {
		elem.next.prev = elem.prev
	} else {
		c.tail = elem.prev
	}

	elem.next = nil
	elem.prev = nil
	c.listLen--
}

func (c *Cache[K, V]) moveToFront(elem *element[K, V]) {
	if c.head == elem {
		return
	}

	c.removeElement(elem)
	c.pushFront(elem)
}

func (c *Cache[K, V]) pushFront(elem *element[K, V]) {
	elem.next = c.head

	if c.head != nil {
		c.head.prev = elem
	}

	c.head = elem

	if c.tail == nil {
		c.tail = elem
	}

	c.listLen++
}
