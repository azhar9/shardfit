package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/azhar9/shardfit/internal/adapter"
	"github.com/azhar9/shardfit/internal/timings"
)

type reportFlags struct {
	junitXML   []string
	timingsRef string
	timingsOut string
	history    int
	pruneAfter int
}

func newReportCmd(a adapter.Adapter) *cobra.Command {
	f := &reportFlags{}
	cmd := &cobra.Command{
		Use:          "report",
		Short:        "Merge JUnit XML durations into the timings store",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReport(a, f)
		},
	}
	cmd.Flags().StringSliceVar(&f.junitXML, "junit-xml", nil, "JUnit XML file (repeatable; globs expanded)")
	cmd.Flags().StringVar(&f.timingsRef, "timings", "", "existing timings store path or URL")
	cmd.Flags().StringVar(&f.timingsOut, "timings-out", "", "write path (defaults to --timings when local)")
	cmd.Flags().IntVar(&f.history, "history", 5, "durations per test to keep")
	cmd.Flags().IntVar(&f.pruneAfter, "prune-after", 30, "drop tests unseen for this many days")
	_ = cmd.MarkFlagRequired("junit-xml")
	return cmd
}

func runReport(a adapter.Adapter, f *reportFlags) error {
	files, err := expandXML(f.junitXML)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no JUnit XML files found for %v", f.junitXML)
	}
	// destination first: fail before any network or parse work
	out := f.timingsOut
	if out == "" {
		out = f.timingsRef
	}
	if out == "" {
		return fmt.Errorf("no destination: pass --timings (local path) or --timings-out")
	}
	if isURL(out) {
		return fmt.Errorf("cannot write timings to a URL (%s); use --timings-out with a local path", out)
	}
	if f.pruneAfter < 1 {
		return fmt.Errorf("--prune-after must be >= 1, got %d", f.pruneAfter)
	}
	store, err := timings.Load(f.timingsRef)
	if err != nil {
		return err
	}
	durations := map[string]int64{}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read %s: %w", file, err)
		}
		perFile, err := a.ParseDurations(data)
		if err != nil {
			return fmt.Errorf("parse %s: %w", file, err)
		}
		for id, d := range perFile {
			durations[id] += d
		}
	}
	today := time.Now()
	merged, added := store.Merge(durations, today, f.history)
	pruned := store.Prune(today, f.pruneAfter)
	if err := store.Save(out); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "merged %d durations from %d files, added %d new tests, pruned %d stale\n", merged+added, len(files), added, pruned)
	return nil
}

func isURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https")
}

// expandXML glob-expands each argument and dedupes by resolved path. A glob
// that matches nothing contributes nothing (caller errors when the result is
// empty); a literal path is kept as-is so its read error is clear.
func expandXML(args []string) ([]string, error) {
	seen := map[string]bool{}
	var out []string
	for _, arg := range args {
		matches, err := filepath.Glob(arg)
		if err != nil {
			return nil, fmt.Errorf("bad glob %q: %w", arg, err)
		}
		if len(matches) == 0 {
			if strings.ContainsAny(arg, "*?[") {
				continue
			}
			matches = []string{arg}
		}
		for _, m := range matches {
			abs, err := filepath.Abs(m)
			if err != nil {
				return nil, err
			}
			if !seen[abs] {
				seen[abs] = true
				out = append(out, m)
			}
		}
	}
	return out, nil
}
