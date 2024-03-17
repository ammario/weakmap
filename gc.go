package weakmap

import (
	"runtime"
	"runtime/metrics"
)

// Refer to https://go.dev/doc/gc-guide.
// And, https://pkg.go.dev/runtime/metrics#Value.

type memStats struct {
	samples []metrics.Sample

	total        uint64
	heapReleased uint64
	used         uint64
	numCycles    uint64
}

// pressure computes a number between 0 and 1 that represents the memory pressure.
// "total" is selected as memory mapped to the process. In the future we may
// want to use to the minimum of the system memory or GOMEMLIMIT, as Go will
// ask for far less memory than it actually needs.
func (m *memStats) pressure() float64 {
	return (float64(m.used)) / (float64(m.total))
}

func (m *memStats) read() {
	if m.samples == nil {
		m.samples = []metrics.Sample{
			{Name: "/memory/classes/total:bytes"},
			{Name: "/memory/classes/heap/released:bytes"},
			{Name: "/memory/classes/heap/objects:bytes"},
			{Name: "/gc/cycles/total:gc-cycles"},
		}
	}
	metrics.Read(m.samples)
	m.total = m.samples[0].Value.Uint64()
	m.heapReleased = m.samples[1].Value.Uint64()
	m.used = m.samples[2].Value.Uint64()
	m.numCycles = m.samples[3].Value.Uint64()
}

// initFinalizerChain sets up a finalizer to run the GC, and then recreates itself
// upon ever GC.
func (l *Map[K, V]) initFinalizerChain() {
	// This sentinel value escapes into the heap and is used to "hook" into
	// the GC.
	sentinel := allocSentinel()
	defer runtime.KeepAlive(sentinel)

	runtime.SetFinalizer(sentinel, func(s *gcSentinel) {
		l.mu.Lock()
		defer l.mu.Unlock()

		l.gc(s)
		// IMPORTANT: do not create a finalizer when the map is empty
		// to avoid an infinite leak of the sentinel object.
		if len(l.index) > 0 {
			l.initFinalizerChain()
		}
	})
}

func (l *Map[K, V]) gc(s *gcSentinel) {
	l.lastSentinel = s
	l.gcMemStats.read()

	// memPressure is used to determine the probability of evicting an entry.
	var (
		memPressure = l.gcMemStats.pressure()
		toEvict     = memPressure * float64(len(l.index))
	)

	// fmt.Printf("pressure: %0.2f, toEvict: %0.2f\n", memPressure, toEvict)

	for i := 0; i <= int(toEvict); i++ {
		last := l.lruList.Tail()
		if last == nil {
			return
		}
		l.delete(last.Data.key)
	}
}
