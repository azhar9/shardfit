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
	tests := []Test{{ID: "known"}, {ID: "fresh"}}
	expected := map[string]int64{"known": 100}
	buckets, err := Partition(tests, expected, 90, 2, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(buckets[0]) != 1 || len(buckets[1]) != 1 {
		t.Fatalf("each bucket should hold one test, got %v", buckets)
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

func TestPartitionRejectsZeroBuckets(t *testing.T) {
	_, err := Partition([]Test{{ID: "a"}}, nil, 0, 0, false)
	if err == nil {
		t.Fatal("want error for 0 buckets")
	}
}
