// Package e2e validates the core claim: a timing-informed split balances a
// fake suite better than an uninformed (cold-start) split.
package e2e

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/azhar9/shardfit/internal/split"
	"github.com/azhar9/shardfit/internal/timings"
)

func bucketTotals(buckets [][]string, trueDur map[string]int64) []int64 {
	totals := make([]int64, len(buckets))
	for i, ids := range buckets {
		for _, id := range ids {
			totals[i] += trueDur[id]
		}
	}
	return totals
}

func imbalance(buckets [][]string, trueDur map[string]int64) float64 {
	totals := bucketTotals(buckets, trueDur)
	var max, min int64
	for i, v := range totals {
		if i == 0 || v > max {
			max = v
		}
		if i == 0 || v < min {
			min = v
		}
	}
	if max == 0 {
		return 0
	}
	return 100 * float64(max-min) / float64(max)
}

func TestTimingInformedSplitBeatsUninformed(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	const n = 50
	tests := make([]split.Test, n)
	trueDur := map[string]int64{}
	for i := range tests {
		id := fmt.Sprintf("test_%02d", i)
		tests[i] = split.Test{ID: id}
		trueDur[id] = 1 + int64(rng.Intn(200)) // 1..200 ms
	}
	// uninformed baseline: cold start, equal estimates
	base, err := split.Partition(tests, nil, 50, 5, false)
	if err != nil {
		t.Fatal(err)
	}
	baseImb := imbalance(base, trueDur)

	// "run" the suite once: report folds the real durations into the store
	store, _ := timings.Load("")
	store.Merge(trueDur, time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC), 5)

	expected := map[string]int64{}
	var known []int64
	for _, ts := range tests {
		if e, ok := store.ExpectedFor(ts.ID, 5); ok {
			expected[ts.ID] = e
			known = append(known, e)
		}
	}
	informed, err := split.Partition(tests, expected, timings.Median(known), 5, false)
	if err != nil {
		t.Fatal(err)
	}
	informedImb := imbalance(informed, trueDur)
	if informedImb >= baseImb {
		t.Fatalf("informed split did not improve: baseline %.1f%%, informed %.1f%%", baseImb, informedImb)
	}
	t.Logf("baseline %.1f%%, informed %.1f%%", baseImb, informedImb)
}
