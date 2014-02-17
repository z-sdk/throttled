package throttled

import (
	"net/http"
	"runtime"
	"testing"
	"time"

	"github.com/PuerkitoBio/boom/commands"
)

func TestMemStats(t *testing.T) {
	cases := []struct {
		n     int
		c     int
		gc    uint32
		total uint64
		rate  time.Duration
	}{
		0: {1000, 10, 3, 0, 0},
		1: {200, 10, 0, 600000, 0},
		2: {500, 10, 2, 555555, 10 * time.Millisecond},
	}
	var mem runtime.MemStats
	for i, c := range cases {
		// Setup the stats handler
		st := &stats{}
		runtime.ReadMemStats(&mem)
		limit := new(runtime.MemStats)
		if c.gc > 0 {
			limit.NumGC = mem.NumGC + c.gc
		}
		if c.total > 0 {
			limit.TotalAlloc = mem.TotalAlloc + c.total
		}
		// Create the throttler
		th := MemStats(limit, c.rate)
		th.DroppedHandler = http.HandlerFunc(st.DroppedHTTP)
		// Run the test
		b := commands.Boom{
			Req:    &commands.ReqOpts{},
			N:      c.n,
			C:      c.c,
			Output: "quiet",
		}
		rpts := runTest(th.Throttle(st), b)
		// Assert results
		assertStats(t, i, st, rpts)
		assertMem(t, i, limit)
	}
}

func assertMem(t *testing.T, ix int, limit *runtime.MemStats) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	if mem.NumGC < limit.NumGC {
		t.Errorf("%d: expected gc to be at least %d, got %d", ix, limit.NumGC, mem.NumGC)
	}
	if mem.TotalAlloc < limit.TotalAlloc {
		t.Errorf("%d: expected total alloc to be at least %dKb, got %dKb", ix, limit.TotalAlloc/1024, mem.TotalAlloc/1024)
	}
}

func BenchmarkReadMemStats(b *testing.B) {
	var mem runtime.MemStats
	for i := 0; i < b.N; i++ {
		runtime.ReadMemStats(&mem)
	}
}
