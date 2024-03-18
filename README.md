# weakmap
[![Go Reference](https://pkg.go.dev/badge/github.com/ammario/weakmap.svg)](https://pkg.go.dev/github.com/ammario/weakmap@main)

Package `weakmap` implements a weak map for Go without `unsafe` or pointer magic.
Instead, it uses finalizers to hook into garbage collection cycles and evicts
old entries based on a combination of a least-recently-used (LRU) policy
and memory pressure reported by the runtime.

The primary use case for this type of weak map is memory-caching. Compared to a
traditional in-memory caches, you don't have to think about deadlines or background cleanup tasks.
You can carelessly throw stuff into the weak map and let the GC take care of the rest.

Install:
```
go get github.com/ammario/weakmap@main
```


> [!WARNING]
> `weakmap` relies on `runtime.SetFinalizer` and, specifically, that
> objects larger than 16 bytes will not be batched into shared allocation
> slots. While this behavior is documented, whether the Go Authors consider
> it apart of the compatibility promise is dubious at best.

## Basic Usage

The API should be familiar to anyone that's used a map:

```go
// The default value is good to use.
m := weakmap.Map[string, int]{}
m.Set("pi", 3.14)

// 3.14, true
v, ok := m.Get("pi")

m.Delete("pi")

// It's now gone!
```

The map's operations are already protected by a mutex and are safe for concurrent
use.

## Eviction

Cache eviction occurs automatically when the GC runs. The number of removed
entries is proportional to the program's memory pressure.

Memory pressure is defined as the ratio of [`runtime/metrics`](https://pkg.go.dev/runtime/metrics) `/memory/classes/heap/objects:bytes` to `/memory/classes/total:bytes`. The Map likely needs a tuning
parameter to adjust sensitivity but I'm not yet sure what that should be.

See [`gc.go`](./gc.go) for the implementation details.

## Testing
You may run ./example/gctest to see how the map behaves under different
memory conditions.

For example:
```bash
$ go run ./example/gctest/ -memlimit 100000000 -allocsize 1000000 -pause 10ms
```

```text
allocating at a rate of 100 MB/s to a memory limit of 100 MB
Map Size  Total Sets Mem Alloc   Next GC   GC Runs
13        99         43 MB (42%) 67 MB     7
8         199       49 MB (48%) 85 MB     10
29        299       88 MB (87%) 91 MB     12
36        399       87 MB (86%) 91 MB     15
35        499       75 MB (74%) 85 MB     18
27        599       55 MB (54%) 59 MB     21
34        699       85 MB (84%) 91 MB     24
33        799       73 MB (72%) 85 MB     27
26        900       54 MB (53%) 59 MB     30
32        999       83 MB (82%) 91 MB     33
```

## Further Reading

* See https://github.com/golang/go/issues/43615 for the state of the debate of
weak references in Go.