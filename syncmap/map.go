// Copyright (c) 2020-2024 Richard Cooper
//
// This file is a part of quenbyako/ext package. See
// https://github.com/quenbyako/ext/blob/master/LICENSE for details

package syncmap

import (
	"sync"
)

// Map is like a Go map[any]any but is safe for concurrent use
// by multiple goroutines without additional locking or coordination.
// Loads, stores, and deletes run in amortized constant time.
//
// The Map type is specialized. Most code should use a plain Go map instead,
// with separate locking or coordination, for better type safety and to make it
// easier to maintain other invariants along with the map content.
//
// The Map type is optimized for two common use cases: (1) when the entry for a given
// key is only ever written once but read many times, as in caches that only grow,
// or (2) when multiple goroutines read, write, and overwrite entries for disjoint
// sets of keys. In these two cases, use of a Map may significantly reduce lock
// contention compared to a Go map paired with a separate [Mutex] or [RWMutex].
//
// The zero Map is empty and ready for use. A Map must not be copied after first use.
//
// In the terminology of [the Go memory model], Map arranges that a write operation
// “synchronizes before” any read operation that observes the effect of the write, where
// read and write operations are defined as follows.
// [Map.Load], [Map.LoadAndDelete], [Map.LoadOrStore], [Map.Swap], [Map.CompareAndSwap],
// and [Map.CompareAndDelete] are read operations;
// [Map.Delete], [Map.LoadAndDelete], [Map.Store], and [Map.Swap] are write operations;
// [Map.LoadOrStore] is a write operation when it returns loaded set to false;
// [Map.CompareAndSwap] is a write operation when it returns swapped set to true;
// and [Map.CompareAndDelete] is a write operation when it returns deleted set to true.
//
// [the Go memory model]: https://go.dev/ref/mem
type Map[K comparable, V any] struct{ m sync.Map }

// Delete deletes the value for a key.
func (m *Map[K, V]) Delete(key K) { m.m.Delete(key) }

// Store sets the value for a key.
func (m *Map[K, V]) Store(key K, value V) { m.m.Store(key, value) }

// Load returns the value stored in the map for a key, or nil if no value is
// present. The ok result indicates whether value was found in the map.
func (m *Map[K, V]) Load(key K) (value V, ok bool) {
	if v, ok := m.m.Load(key); ok {
		return v.(V), true
	}

	return value, false
}

// LoadAndDelete deletes the value for a key, returning the previous value if
// any. The loaded result reports whether the key was present.
func (m *Map[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	if v, loaded := m.m.LoadAndDelete(key); loaded {
		return v.(V), true
	}

	return value, false
}

// LoadOrStore returns the existing value for the key if present. Otherwise, it
// stores and returns the given value. The loaded result is true if the value
// was loaded, false if stored.
func (m *Map[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	v, loaded := m.m.LoadOrStore(key, value)
	return v.(V), loaded
}

// Range calls f sequentially for each key and value present in the map. If f
// returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently (including by f), Range may reflect any
// mapping for that key from any point during the Range call. Range does not
// block other methods on the receiver; even f itself may call any method on m.
//
// Range may be O(N) with the number of elements in the map even if f returns
// false after a constant number of calls.
func (m *Map[K, V]) Range(f func(key K, value V) bool) {
	m.m.Range(func(key, value any) bool { return f(key.(K), value.(V)) })
}
