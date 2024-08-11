# Ext

## `user`

Default `os/user` doesn't provide any interface to work with other systems. So
this package creates unified interface for different implementations of getting
POSIX users and groups (you can implement ACL logic even inside your web app!)

## `set`

Set is a missed collection for Go. This package provides a set data structure
that can be classic non-threadsafe version like classic `map[T]struct{}` or
threadsafe version with `sync.Map` under the hood.

Set provides not just a `map[T]struct{}` type, but also a set of methods to work
with sets like `Union`, `Intersection`, `Difference`, `SymmetricDifference`.

## `span`

Span provides a simple way to work with intervals of any time, from `time.Time`
to `semver.Version` types.
