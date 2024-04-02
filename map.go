package weakmap

import (
	"sync"

	"github.com/ammario/weakmap/doublelist"
)

// dataWithKey bundles data with its reference key.
// This structure allows for reverse lookup from the doubly-linked list to the index.
type dataWithKey[K comparable, V any] struct {
	data V
	key  K
}

// Map implements a LRU weak map safe for concurrent use.
// The zero value is an empty map ready for use.
//
// When the Go GC runs, the oldest entries proportional to memory
// pressure are evicted.
type Map[K comparable, V any] struct {
	mu sync.Mutex

	index map[K]*doublelist.Node[dataWithKey[K, V]]

	// lruList contains entries in order of least-recently-used to most-recently-used.
	lruList *doublelist.List[dataWithKey[K, V]]

	gcMemStats   memStats
	lastSentinel *gcSentinel

	cost int
	// Coster is a function that returns the cost of a value,
	// policed by MaxCost. If left nil, Coster becomes a function
	// that always returns 1.
	Coster func(V) int
	// MaxCost is the maximum cost of the cache.
	// When a Set would cause the cost to exceed MaxCost,
	// the least recently used entries are evicted until
	// the cost is below MaxCost.
	MaxCost int
}

func (m *Map[K, V]) initOnce() {
	if m.index != nil {
		return
	}
	m.index = make(map[K]*doublelist.Node[dataWithKey[K, V]])
	m.lruList = &doublelist.List[dataWithKey[K, V]]{}
	if m.Coster == nil {
		m.Coster = func(V) int { return 1 }
	}
}

type gcSentinel struct {
	// Get around the 16-byte allocation batching
	// to be extra-confident the finalizer runs.
	_ [24]byte
}

func allocSentinel() *gcSentinel {
	return &gcSentinel{}
}

func (l *Map[K, V]) delete(key K) {
	node, ok := l.index[key]
	if !ok {
		return
	}
	l.lruList.Pop(node)
	l.cost -= l.Coster(node.Data.data)
	delete(l.index, key)
}

// Delete removes an entry from the cache.
func (l *Map[K, V]) Delete(key K) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.initOnce()

	l.delete(key)
}

// Set adds a new value to the cache.
//
// Set may also be used to bump an existing value to the top of the cache.
func (l *Map[K, V]) Set(key K, v V) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.initOnce()

	// Remove existing key if it exists.
	l.delete(key)

	l.index[key] = l.lruList.Append(
		dataWithKey[K, V]{
			data: v,
			key:  key,
		},
	)
	l.cost += l.Coster(v)

	// Set up the finalizer chain upon first entry.
	if len(l.index) == 1 {
		l.initFinalizerChain()
	}

	l.evictOverages()
}

func (l *Map[K, V]) evictOverages() {
	if l.MaxCost == 0 {
		// Disabled.
		return
	}
	for l.cost > l.MaxCost {
		last := l.lruList.Tail()
		if last == nil {
			// No data left to evictOverages. Avoid looping forever.
			return
		}
		l.delete(last.Data.key)
	}
}

func (l *Map[K, V]) get(key K) (v V, exists bool) {
	node, exists := l.index[key]
	if !exists {
		return v, false
	}

	// Bump value to top.
	l.lruList.Pop(node)
	l.index[key] = l.lruList.Append(node.Data)
	return node.Data.data, true
}

// Get retrieves a value from the cache, if it exists.
func (l *Map[K, V]) Get(key K) (v V, exists bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.initOnce()

	return l.get(key)
}

func (l *Map[K, V]) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.initOnce()

	return len(l.index)
}

// Cost returns the current cost of the cache.
func (l *Map[K, V]) Cost() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.initOnce()

	return l.cost
}

// Do is a helper that retrieves a value from the cache, if it exists, and
// calls the provided function to compute the value if it does not.
func (l *Map[K, V]) Do(key K, fn func() (V, error)) (V, error) {
	v, ok := l.Get(key)
	if ok {
		return v, nil
	}

	v, err := fn()
	if err != nil {
		return v, err
	}

	l.Set(key, v)
	return v, nil
}
