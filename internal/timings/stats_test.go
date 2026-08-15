package timings

import "testing"

func TestWeightedMedianRecentDominates(t *testing.T) {
	if got := WeightedMedian([]int64{100, 200}); got != 200 {
		t.Fatalf("WeightedMedian([100 200]) = %d, want 200", got)
	}
}

func TestWeightedMedianEmpty(t *testing.T) {
	if got := WeightedMedian(nil); got != 0 {
		t.Fatalf("WeightedMedian(nil) = %d, want 0", got)
	}
}

func TestMedian(t *testing.T) {
	if got := Median([]int64{1, 2, 3}); got != 2 {
		t.Fatalf("Median([1 2 3]) = %d, want 2", got)
	}
	if got := Median([]int64{1, 2, 3, 4}); got != 2 {
		t.Fatalf("Median([1 2 3 4]) = %d, want 2", got)
	}
}

func TestPercentile(t *testing.T) {
	if got := Percentile([]int64{1, 2, 3, 4, 5}, 25); got != 2 {
		t.Fatalf("Percentile(25) = %d, want 2", got)
	}
	if got := Percentile([]int64{1, 2, 3, 4, 5}, 99); got != 5 {
		t.Fatalf("Percentile(99) = %d, want 5", got)
	}
}

func TestCoefficientOfVariation(t *testing.T) {
	if got := CoefficientOfVariation([]int64{1, 1, 1}); got != 0 {
		t.Fatalf("CV of constants = %v, want 0", got)
	}
	if got := CoefficientOfVariation(nil); got != 0 {
		t.Fatalf("CV of nil = %v, want 0", got)
	}
}

func TestExpectedDurationStable(t *testing.T) {
	if got := ExpectedDuration([]int64{100, 100, 100}); got != 100 {
		t.Fatalf("ExpectedDuration stable = %d, want 100", got)
	}
}

func TestExpectedDurationDampensFlaky(t *testing.T) {
	// one 500ms outlier among 100ms runs: CV > 0.5 → dampened toward P25 (100)
	if got := ExpectedDuration([]int64{100, 100, 100, 100, 500}); got != 100 {
		t.Fatalf("ExpectedDuration flaky = %d, want 100", got)
	}
}

func TestExpectedDurationEmpty(t *testing.T) {
	if got := ExpectedDuration(nil); got != 0 {
		t.Fatalf("ExpectedDuration(nil) = %d, want 0", got)
	}
}
