package split

import (
	"reflect"
	"strings"
	"testing"
)

func loads(buckets [][]string, expected map[string]int64) []int64 {
	out := make([]int64, len(buckets))
	for i, b := range buckets {
		for _, id := range b {
			out[i] += expected[id]
		}
	}
	return out
}

func TestPartitionBalances(t *testing.T) {
	tests := []Test{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}}
	expected := map[string]int64{"a": 10, "b": 6, "c": 5, "d": 4}
	buckets, err := Partition(tests, expected, 0, 2, false)
	if err != nil {
		t.Fatal(err)
	}
	got := loads(buckets, expected)
	// LPT: sorted desc a=10,b=6,c=5,d=4 → a→0(10), b→1(6), c→1(11), d→0(14)
	if !reflect.DeepEqual(got, []int64{14, 11}) {
		t.Fatalf("loads = %v, want [14 11]", got)
	}
}

func TestPartitionDeterministic(t *testing.T) {
	tests := []Test{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}}
	expected := map[string]int64{"a": 10, "b": 6, "c": 5, "d": 4}
	first, err := Partition(tests, expected, 0, 2, false)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		again, err := Partition(tests, expected, 0, 2, false)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(first, again) {
			t.Fatalf("partition not deterministic:\n%v\n%v", first, again)
		}
	}
}

func TestPartitionUsesUnknownEstimate(t *testing.T) {
	tests := []Test{{ID: "known"}, {ID: "fresh1"}, {ID: "fresh2"}}
	expected := map[string]int64{"known": 100}
	buckets, err := Partition(tests, expected, 80, 2, false)
	if err != nil {
		t.Fatal(err)
	}
	got := loads(buckets, map[string]int64{"known": 100, "fresh1": 80, "fresh2": 80})
	if !reflect.DeepEqual(got, []int64{100, 160}) {
		t.Fatalf("loads = %v, want [100 160] (unknown estimate must count)", got)
	}
}

func TestPartitionSortsLongestFirst(t *testing.T) {
	// input deliberately unsorted: plain input-order greedy yields [6 14],
	// LPT (sorted desc) yields [10 10]
	tests := []Test{{ID: "a"}, {ID: "b"}, {ID: "c"}}
	expected := map[string]int64{"a": 6, "b": 4, "c": 10}
	buckets, err := Partition(tests, expected, 0, 2, false)
	if err != nil {
		t.Fatal(err)
	}
	if got := loads(buckets, expected); !reflect.DeepEqual(got, []int64{10, 10}) {
		t.Fatalf("loads = %v, want [10 10] (LPT requires sorting longest-first)", got)
	}
}

func TestPartitionTooManyBuckets(t *testing.T) {
	tests := []Test{{ID: "a"}, {ID: "b"}}
	_, err := Partition(tests, nil, 0, 3, false)
	if err == nil || !strings.Contains(err.Error(), "more buckets than tests") {
		t.Fatalf("err = %v, want 'more buckets than tests'", err)
	}
}

func TestPartitionGroupByFile(t *testing.T) {
	tests := []Test{
		{ID: "f1::t1", File: "f1"},
		{ID: "f1::t2", File: "f1"},
		{ID: "f2::t3", File: "f2"},
	}
	expected := map[string]int64{"f1::t1": 40, "f1::t2": 40, "f2::t3": 30}
	buckets, err := Partition(tests, expected, 0, 2, true)
	if err != nil {
		t.Fatal(err)
	}
	var f1Bucket int
	for i, b := range buckets {
		for _, id := range b {
			if strings.HasPrefix(id, "f1::") {
				f1Bucket = i
			}
		}
	}
	for _, id := range buckets[f1Bucket] {
		if !strings.HasPrefix(id, "f1::") {
			t.Fatalf("file group split across buckets: %v", buckets)
		}
	}
}

func TestPartitionGroupByFileDeterministic(t *testing.T) {
	tests := []Test{
		{ID: "f1::t1", File: "f1"},
		{ID: "f2::t2", File: "f2"},
		{ID: "f1::t3", File: "f1"},
		{ID: "f3::t4", File: "f3"},
	}
	expected := map[string]int64{"f1::t1": 10, "f2::t2": 20, "f1::t3": 30, "f3::t4": 5}
	first, err := Partition(tests, expected, 0, 2, true)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		again, err := Partition(tests, expected, 0, 2, true)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(first, again) {
			t.Fatalf("grouped partition not deterministic:\n%v\n%v", first, again)
		}
	}
}

func TestPartitionRejectsZeroBuckets(t *testing.T) {
	_, err := Partition([]Test{{ID: "a"}}, nil, 0, 0, false)
	if err == nil {
		t.Fatal("want error for 0 buckets")
	}
}

func TestPartitionSpreadsZeroWeights(t *testing.T) {
	// cold start with no history: every estimate is 0, and loads never
	// differ — buckets must still round-robin on test count, not pile
	// everything into bucket 0
	tests := []Test{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}}
	buckets, err := Partition(tests, nil, 0, 2, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(buckets[0]) != 2 || len(buckets[1]) != 2 {
		t.Fatalf("buckets = %v, want 2 and 2", buckets)
	}
}
