package e2e

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/azhar9/shardfit/internal/split"
	"github.com/azhar9/shardfit/internal/timings"
)

// TestScaleTenThousand proves the operational envelope: 10k tests through
// the real store and partition path, balanced, deterministic, and fast.
func TestScaleTenThousand(t *testing.T) {
	if testing.Short() {
		t.Skip("scale test")
	}
	rng := rand.New(rand.NewSource(7))
	const n, buckets = 10_000, 50
	start := time.Now()

	tests := make([]split.Test, n)
	trueDur := map[string]int64{}
	for i := range tests {
		id := fmt.Sprintf("test_%05d", i)
		tests[i] = split.Test{ID: id}
		trueDur[id] = 1 + int64(rng.Intn(50))*int64(rng.Intn(50)) // 1..2401ms, skewed
	}
	store, _ := timings.Load("")
	store.Merge(trueDur, time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC), 5)

	expected := map[string]int64{}
	for _, ts := range tests {
		e, ok := store.ExpectedFor(ts.ID, 5)
		if !ok {
			t.Fatalf("missing expected duration for %s", ts.ID)
		}
		expected[ts.ID] = e
	}
	first, err := split.Partition(tests, expected, timings.Median([]int64{500}), buckets, false)
	if err != nil {
		t.Fatal(err)
	}
	imb := imbalance(first, trueDur)
	if imb > 5 {
		t.Fatalf("imbalance %.2f%% exceeds 5%% at 10k scale", imb)
	}
	// determinism at scale
	again, err := split.Partition(tests, expected, timings.Median([]int64{500}), buckets, false)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%v", first) != fmt.Sprintf("%v", again) {
		t.Fatal("partition not deterministic at 10k scale")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("10k pipeline took %v, want < 5s", elapsed)
	}
	t.Logf("10k tests, %d buckets: imbalance %.2f%%, %v", buckets, imb, time.Since(start))
}
