package e2e

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/azhar9/shardfit/internal/split"
	"github.com/azhar9/shardfit/internal/timings"
)

// runScale proves the operational envelope at a given size: n synthetic
// tests through the real store and partition path, balanced, deterministic,
// and inside the wall-clock budget.
func runScale(t *testing.T, n, buckets int, budget time.Duration) {
	t.Helper()
	rng := rand.New(rand.NewSource(7))
	start := time.Now()

	tests := make([]split.Test, n)
	trueDur := map[string]int64{}
	for i := range tests {
		id := fmt.Sprintf("test_%06d", i)
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
	if imb := imbalance(first, trueDur); imb > 5 {
		t.Fatalf("imbalance %.2f%% exceeds 5%% at %d scale", imb, n)
	}
	// determinism at scale
	again, err := split.Partition(tests, expected, timings.Median([]int64{500}), buckets, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, again) {
		t.Fatalf("partition not deterministic at %d scale", n)
	}
	if elapsed := time.Since(start); elapsed > budget {
		t.Fatalf("%d pipeline took %v, want < %v", n, elapsed, budget)
	}
	t.Logf("%d tests, %d buckets: imbalance %.2f%%, %v", n, buckets, imbalance(first, trueDur), time.Since(start))
}

func TestScaleTenThousand(t *testing.T) {
	if testing.Short() {
		t.Skip("scale test")
	}
	runScale(t, 10_000, 50, 5*time.Second)
}

func TestScaleHundredThousand(t *testing.T) {
	if testing.Short() {
		t.Skip("scale test")
	}
	runScale(t, 100_000, 100, 10*time.Second)
}
