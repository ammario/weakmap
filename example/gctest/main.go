package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"runtime/debug"
	"sync/atomic"
	"text/tabwriter"
	"time"

	"github.com/ammario/weakmap"
	"github.com/dustin/go-humanize"
)

func main() {
	var (
		memlimit   int
		allocSize  int
		allocPause time.Duration
	)

	flag.IntVar(&memlimit, "memlimit", 1_000_000, "memory limit")
	flag.DurationVar(&allocPause, "pause", 1*time.Millisecond, "pause between allocations")
	flag.IntVar(&allocSize, "allocsize", 1_000, "allocation size")

	flag.Parse()

	debug.SetMemoryLimit(int64(memlimit))

	fmt.Printf(
		"allocating at a rate of %v/s to a memory limit of %v\n",
		humanize.Bytes(uint64(allocSize*int(time.Second/allocPause))),
		humanize.Bytes(uint64(memlimit)),
	)

	var (
		numSets int64
		m       = weakmap.Map[int, []byte]{}
	)
	go func() {
		tw := tabwriter.NewWriter(os.Stdout, 10, 4, 1, ' ', 0)
		ticker := time.NewTicker(1 * time.Second)
		fmt.Fprintf(tw, "Map Size\tTotal Sets\tMem Alloc\tNext GC\tGC Runs\n")
		var (
			gcstats  debug.GCStats
			memStats runtime.MemStats
		)
		for range ticker.C {
			debug.ReadGCStats(&gcstats)
			runtime.ReadMemStats(&memStats)

			fmt.Fprintf(
				tw, "%v\t%v\t%v\t%v\t%v\n",
				m.Len(),
				atomic.LoadInt64(&numSets),
				fmt.Sprintf("%v (%v%%)", humanize.Bytes(uint64(memStats.Alloc)), (memStats.Alloc*100)/uint64(memlimit)),
				humanize.Bytes(uint64(memStats.NextGC)),
				gcstats.NumGC,
			)
			tw.Flush()
		}
	}()

	allocTicker := time.NewTicker(allocPause)
	for ; ; atomic.AddInt64(&numSets, 1) {
		<-allocTicker.C
		m.Set(rand.Int(), make([]byte, allocSize))
	}
}
