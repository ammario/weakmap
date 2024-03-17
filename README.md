# weakmap
[![Go Reference](https://pkg.go.dev/badge/github.com/ammario/weakmap.svg)](https://pkg.go.dev/github.com/ammario/weakmap@main)

Package `weakmap` implements a weak map for Go without `unsafe` or pointer magic.
Instead, it uses finalizers to hook into garbage collection cycles and evicts
old entries based on a combination of the least-recently-used (LRU) policy
and memory pressure reported by the runtime.

```
go get github.com/ammario/weakmap@master
```

See https://github.com/golang/go/issues/43615 for the state of the debate of
weak references in Go.

## Basic Usage

The API should be familiar to anyone that's used a map:

```go
// The default value is good to use.
m := weakmap.Map[string, int]{}
m.Set("pi", 3.14)

// 3.14, true
v, ok := m.Get("pi")
```

## Eviction

Cache eviction occurs automatically when the GC runs. The number of removed
entries is proportional to the program's memory pressure. Memory pressure
is defined as the ratio of `/memory/classes/heap/objects:bytes` to
`/memory/classes/total:bytes`.


