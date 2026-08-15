// Package split implements deterministic test partitioning.
package split

import (
	"fmt"
	"sort"
)

// Test is one discoverable unit. File is empty when the framework has no
// file concept (see --group-by file).
type Test struct {
	ID   string
	File string
}

type item struct {
	key      string
	ids      []string
	expected int64
}

// Partition assigns tests to n buckets by expected duration: longest first
// into the currently least-loaded bucket, ties broken by lowest bucket
// index. When groupByFile is true, tests sharing a non-empty File stay in
// one bucket and their expected durations are summed. Tests with no entry in
// expected get unknownEstimate. Returns one id slice per bucket, in
// assignment order. Fully deterministic for identical input.
func Partition(tests []Test, expected map[string]int64, unknownEstimate int64, n int, groupByFile bool) ([][]string, error) {
	if n < 1 {
		return nil, fmt.Errorf("buckets must be >= 1, got %d", n)
	}
	items := group(tests, expected, unknownEstimate, groupByFile)
	if len(items) < n {
		return nil, fmt.Errorf("cannot split %d tests or groups into %d buckets: more buckets than tests", len(items), n)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].expected != items[j].expected {
			return items[i].expected > items[j].expected
		}
		return items[i].key < items[j].key
	})
	buckets := make([][]string, n)
	loads := make([]int64, n)
	for _, it := range items {
		b := 0
		for i := 1; i < n; i++ {
			if loads[i] < loads[b] || (loads[i] == loads[b] && len(buckets[i]) < len(buckets[b])) {
				b = i
			}
		}
		buckets[b] = append(buckets[b], it.ids...)
		loads[b] += it.expected
	}
	return buckets, nil
}

func group(tests []Test, expected map[string]int64, unknownEstimate int64, groupByFile bool) []item {
	if !groupByFile {
		items := make([]item, 0, len(tests))
		for _, t := range tests {
			items = append(items, item{key: t.ID, ids: []string{t.ID}, expected: exp(expected, t.ID, unknownEstimate)})
		}
		return items
	}
	byFile := map[string]*item{}
	var order []string // first-seen order keeps grouping deterministic
	for _, t := range tests {
		key := t.File
		if key == "" {
			key = t.ID
		}
		if _, ok := byFile[key]; !ok {
			byFile[key] = &item{key: key}
			order = append(order, key)
		}
		it := byFile[key]
		it.ids = append(it.ids, t.ID)
		it.expected += exp(expected, t.ID, unknownEstimate)
	}
	items := make([]item, 0, len(order))
	for _, k := range order {
		items = append(items, *byFile[k])
	}
	return items
}

func exp(expected map[string]int64, id string, unknown int64) int64 {
	if e, ok := expected[id]; ok {
		return e
	}
	return unknown
}
