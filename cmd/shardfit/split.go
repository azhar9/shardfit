package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/azhar9/shardfit/internal/adapter"
	"github.com/azhar9/shardfit/internal/split"
	"github.com/azhar9/shardfit/internal/timings"
)

type splitFlags struct {
	buckets       int
	timingsRef    string
	input         string
	filter        string
	groupBy       string
	unknownEst    string
	history       int
	outlierCap    string
	outDir        string
	estimateOnly  bool
	maxUnknownPct float64
}

func newSplitCmd(a adapter.Adapter) *cobra.Command {
	f := &splitFlags{}
	cmd := &cobra.Command{
		Use:   "split",
		Short: "Write N duration-balanced bucket files",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSplit(a, f)
		},
	}
	cmd.Flags().IntVarP(&f.buckets, "buckets", "n", 0, "number of buckets (required)")
	cmd.Flags().StringVar(&f.timingsRef, "timings", "", "timings store path or URL (missing = cold start)")
	cmd.Flags().StringVar(&f.input, "input", "", "test list file instead of native discovery (- for stdin)")
	cmd.Flags().StringVar(&f.filter, "filter", "", "runner-native filter, forwarded verbatim to discovery")
	cmd.Flags().StringVar(&f.groupBy, "group-by", "auto", "grouping key: auto, file, or test")
	cmd.Flags().StringVar(&f.unknownEst, "unknown-estimate", "", "duration for tests with no history (default: median of known)")
	cmd.Flags().IntVar(&f.history, "history", 5, "durations per test used from the store")
	cmd.Flags().StringVar(&f.outlierCap, "outlier-cap", "p99", "cap estimates at the global P99: p99 or none")
	cmd.Flags().StringVar(&f.outDir, "out-dir", ".", "directory for bucket files")
	cmd.Flags().BoolVar(&f.estimateOnly, "estimate-only", false, "print per-bucket estimates, write no files")
	cmd.Flags().Float64Var(&f.maxUnknownPct, "max-unknown-pct", 50, "error above this percentage of unknown tests")
	_ = cmd.MarkFlagRequired("buckets")
	return cmd
}

func runSplit(a adapter.Adapter, f *splitFlags) error {
	tests, err := a.Discover(adapter.DiscoverConfig{Input: f.input, Filter: f.filter})
	if err != nil {
		return err
	}
	store, err := timings.Load(f.timingsRef)
	if err != nil {
		return err
	}
	expected := map[string]int64{}
	var known []int64
	for _, t := range tests {
		if e, ok := store.ExpectedFor(t.ID, f.history); ok {
			expected[t.ID] = e
			known = append(known, e)
		}
	}
	switch f.outlierCap {
	case "p99":
		if len(known) > 1 {
			cap99 := timings.Percentile(known, 99)
			for id, e := range expected {
				if e > cap99 {
					expected[id] = cap99
				}
			}
			known = known[:0]
			for _, e := range expected {
				known = append(known, e)
			}
		}
	case "none":
	default:
		return fmt.Errorf("--outlier-cap must be p99 or none, got %q", f.outlierCap)
	}
	unknown := len(tests) - len(expected)
	unknownEst := int64(0)
	if f.unknownEst != "" {
		d, err := time.ParseDuration(f.unknownEst)
		if err != nil {
			return fmt.Errorf("--unknown-estimate: %w", err)
		}
		unknownEst = d.Milliseconds()
	} else if len(known) > 0 {
		unknownEst = timings.Median(known)
	}
	if len(tests) > 0 {
		pct := 100 * float64(unknown) / float64(len(tests))
		// a cold start (empty store) never blocks the pipeline
		if pct > f.maxUnknownPct && len(store.Tests) > 0 {
			return fmt.Errorf("%.0f%% of tests have no history (limit %.0f%%); run a baseline first or raise --max-unknown-pct", pct, f.maxUnknownPct)
		}
		if pct > 20 {
			fmt.Fprintf(os.Stderr, "warning: %.0f%% of tests have no history; estimates may be inaccurate\n", pct)
		}
	}
	switch f.groupBy {
	case "auto", "file", "test":
	default:
		return fmt.Errorf("--group-by must be auto, file, or test, got %q", f.groupBy)
	}
	groupByFile := f.groupBy == "file" || (f.groupBy == "auto" && a.Granularity() == adapter.GranularityFile)
	if groupByFile {
		for _, t := range tests {
			if t.File == "" {
				return fmt.Errorf("--group-by file is not supported by adapter %q (no file info); use --group-by test", a.Name())
			}
		}
	}
	splitTests := make([]split.Test, len(tests))
	for i, t := range tests {
		splitTests[i] = split.Test{ID: t.ID, File: t.File}
	}
	buckets, err := split.Partition(splitTests, expected, unknownEst, f.buckets, groupByFile)
	if err != nil {
		return err
	}
	return writeBuckets(buckets, expected, unknownEst, f)
}

func writeBuckets(buckets [][]string, expected map[string]int64, unknownEst int64, f *splitFlags) error {
	totals := make([]int64, len(buckets))
	for i, ids := range buckets {
		for _, id := range ids {
			if e, ok := expected[id]; ok {
				totals[i] += e
			} else {
				totals[i] += unknownEst
			}
		}
	}
	var max, min int64
	for i, tot := range totals {
		if i == 0 || tot > max {
			max = tot
		}
		if i == 0 || tot < min {
			min = tot
		}
	}
	imbalance := 0.0
	if max > 0 {
		imbalance = 100 * float64(max-min) / float64(max)
	}
	fmt.Fprintf(os.Stderr, "bucket\testimated\tcount\n")
	for i, tot := range totals {
		fmt.Fprintf(os.Stderr, "%d\t%s\t%d\n", i+1, time.Duration(tot*int64(time.Millisecond)).String(), len(buckets[i]))
	}
	fmt.Fprintf(os.Stderr, "imbalance: %.1f%%\n", imbalance)
	if f.estimateOnly {
		return nil
	}
	if err := os.MkdirAll(f.outDir, 0o755); err != nil {
		return fmt.Errorf("create out dir: %w", err)
	}
	for i, ids := range buckets {
		path := filepath.Join(f.outDir, fmt.Sprintf("bucket-%d.txt", i+1))
		data := make([]byte, 0, len(ids)*16)
		for _, id := range ids {
			data = append(data, id+"\n"...)
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
		fmt.Fprintln(os.Stderr, "wrote", path)
	}
	return nil
}
