package maps

// Keys returns the keys of the map m.
// The keys will be in an indeterminate order.
func Keys[M ~map[K]V, K comparable, V any](m M) []K {
	r := make([]K, 0, len(m))
	for k := range m {
		r = append(r, k)
	}
	return r
}

// Values returns the values of the map m.
// The values will be in an indeterminate order.
func Values[M ~map[K]V, K comparable, V any](m M) []V {
	r := make([]V, 0, len(m))
	for _, v := range m {
		r = append(r, v)
	}
	return r
}

// Equal reports whether two maps contain the same key/value pairs.
// Values are compared using ==.
func Equal[M1, M2 ~map[K]V, K, V comparable](m1 M1, m2 M2) bool {
	return EqualFunc(m1, m2, func(a, b V) bool { return a == b })
}

// EqualFunc is like Equal, but compares values using eq.
// Keys are still compared with ==.
func EqualFunc[M1 ~map[K]V1, M2 ~map[K]V2, K comparable, V1, V2 any](m1 M1, m2 M2, eq func(V1, V2) bool) bool {
	if len(m1) != len(m2) {
		return false
	}
	for k, v1 := range m1 {
		if v2, ok := m2[k]; !ok || !eq(v1, v2) {
			return false
		}
	}
	return true
}

// Clear removes all entries from m, leaving it empty.
func Clear[M ~map[K]V, K comparable, V any](m M) {
	for k := range m {
		delete(m, k)
	}
}

// Clone returns a copy of m.  This is a shallow clone:
// the new keys and values are set using ordinary assignment.
func Clone[M ~map[K]V, K comparable, V any](m M) M {
	// Preserve nil in case it matters.
	if m == nil {
		return nil
	}
	r := make(M, len(m))
	for k, v := range m {
		r[k] = v
	}
	return r
}

// Copy copies all key/value pairs in src adding them to dst.
// When a key in src is already present in dst,
// the value in dst will be overwritten by the value associated
// with the key in src.
func Copy[M1 ~map[K]V, M2 ~map[K]V, K comparable, V any](dst M1, src M2) M1 {
	for k, v := range src {
		dst[k] = v
	}

	return dst
}

// DeleteFunc deletes any key/value pairs from m for which del returns true.
func DeleteFunc[M ~map[K]V, K comparable, V any](m M, del func(K, V) bool) M {
	for k, v := range m {
		if del(k, v) {
			delete(m, k)
		}
	}

	return m
}

// GetOne gets single random key and value from map. If length of map is zero, it
// returns zero values and `ok` as false.
//
// Careful! "Random key and value" here DOES NOT mean, that it's
// cryptographically random. Instead, it uses range random ordering from golang
// runtime.
func GetOne[M ~map[K]V, K comparable, V any](m M) (k K, v V, ok bool) {
	for key, value := range m {
		return key, value, true
	}

	return k, v, false
}

// Pop gets single random key and value from map and remove it, if at least one
// key exists. If length of map is zero, it returns zero values and `ok` as
// false.
//
// Careful! "Random key and value" here DOES NOT mean, that it's
// cryptographically random. Instead, it uses range random ordering from golang
// runtime.
func Pop[M ~map[K]V, K comparable, V any](m M) (k K, v V, ok bool) {
	k, v, ok = GetOne(m)
	if ok {
		delete(m, k)
	}
	return k, v, ok
}

// Merge merges map items into base map. If you want to create new map, you can
// provide nil to base.
func Merge[M ~map[K]V, K comparable, V any](base M, maps ...M) M {
	if base == nil {
		base = make(M)
	}

	for _, item := range maps {
		for k, v := range item {
			base[k] = v
		}
	}
	return base
}

// Remap remaps map of one type to different one.
func Remap[M1 ~map[K1]V1, K1, K2 comparable, V1, V2 any](m M1, f func(K1, V1) (K2, V2)) map[K2]V2 {
	res := make(map[K2]V2, len(m))
	for k1, v1 := range m {
		k2, v2 := f(k1, v1)
		res[k2] = v2
	}

	return res
}
