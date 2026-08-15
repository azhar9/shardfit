package timings

import (
	"math"
	"sort"
)

// WeightedMedian returns the median of ds with weights = age rank
// (oldest=1 … newest=len(ds)), so recent runs influence the result more.
func WeightedMedian(ds []int64) int64 {
	if len(ds) == 0 {
		return 0
	}
	if len(ds) == 1 {
		return ds[0]
	}
	type wv struct {
		v int64
		w int64
	}
	vals := make([]wv, len(ds))
	var total int64
	for i, d := range ds {
		w := int64(i + 1)
		vals[i] = wv{d, w}
		total += w
	}
	sort.Slice(vals, func(i, j int) bool { return vals[i].v < vals[j].v })
	target := (total + 1) / 2
	var cum int64
	for _, x := range vals {
		cum += x.w
		if cum >= target {
			return x.v
		}
	}
	return vals[len(vals)-1].v // unreachable
}

// Median returns the plain median.
func Median(ds []int64) int64 {
	if len(ds) == 0 {
		return 0
	}
	s := append([]int64(nil), ds...)
	sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })
	m := len(s) / 2
	if len(s)%2 == 1 {
		return s[m]
	}
	return (s[m-1] + s[m]) / 2
}

// Percentile returns the nearest-rank percentile (p in 0..100).
func Percentile(ds []int64, p int) int64 {
	if len(ds) == 0 {
		return 0
	}
	s := append([]int64(nil), ds...)
	sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })
	return s[(len(s)*p-1)/100] // nearest-rank: ceil(n·p/100) − 1
}

// CoefficientOfVariation is stddev/mean; 0 when mean is 0 or ds is empty.
func CoefficientOfVariation(ds []int64) float64 {
	if len(ds) == 0 {
		return 0
	}
	var sum float64
	for _, d := range ds {
		sum += float64(d)
	}
	mean := sum / float64(len(ds))
	if mean == 0 {
		return 0
	}
	var ss float64
	for _, d := range ds {
		ss += (float64(d) - mean) * (float64(d) - mean)
	}
	return math.Sqrt(ss/float64(len(ds))) / mean
}

// ExpectedDuration estimates one test's runtime: recency-weighted median,
// dampened toward the lower quartile when the history is flaky (CV > 0.5):
// 0.7·wmedian + 0.3·min(wmedian, P25).
func ExpectedDuration(ds []int64) int64 {
	if len(ds) == 0 {
		return 0
	}
	wm := WeightedMedian(ds)
	if CoefficientOfVariation(ds) > 0.5 {
		lo := wm
		if p25 := Percentile(ds, 25); p25 < lo {
			lo = p25
		}
		return int64(0.7*float64(wm) + 0.3*float64(lo))
	}
	return wm
}
