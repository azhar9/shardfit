// Package jest adapts shardfit to jest. Jest runs each test file in its own
// process, so sharding happens at file granularity.
package jest

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/azhar9/shardfit/internal/adapter"
	"github.com/azhar9/shardfit/internal/junitxml"
)

// Adapter implements adapter.Adapter for jest.
type Adapter struct{}

// New returns a new Adapter.
func New() *Adapter { return &Adapter{} }

// Compile-time assertion: Adapter implements adapter.Adapter.
var _ adapter.Adapter = (*Adapter)(nil)

func (a *Adapter) Name() string { return "jest" }

func (a *Adapter) Granularity() adapter.Granularity { return adapter.GranularityFile }

// Discover runs `jest --listTests` (Filter passed through verbatim) and
// relativizes paths against the working directory. An Input list bypasses
// jest entirely.
func (a *Adapter) Discover(cfg adapter.DiscoverConfig) ([]adapter.Test, error) {
	if cfg.Input != "" {
		tests, err := adapter.ReadList(cfg.Input, os.Stdin)
		if err != nil {
			return nil, err
		}
		for i := range tests {
			tests[i].File = tests[i].ID
		}
		return tests, nil
	}
	args := []string{"--listTests"}
	if cfg.Filter != "" {
		args = append(adapter.SplitFilter(cfg.Filter), args...)
	}
	out, err := exec.Command("jest", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("jest --listTests failed: %w (output: %s)", err, out)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get cwd: %w", err)
	}
	var tests []adapter.Test
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		p := strings.TrimSpace(sc.Text())
		if p == "" {
			continue
		}
		if rel, err := filepath.Rel(cwd, p); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			p = rel
		}
		id := filepath.ToSlash(p)
		tests = append(tests, adapter.Test{ID: id, File: id})
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read jest output: %w", err)
	}
	return tests, nil
}

// ParseDurations sums per-testcase durations by file. jest-junit's default
// classNameTemplate is the test file path; non-path classnames (custom
// templates) fall back to "classname > name".
func (a *Adapter) ParseDurations(data []byte) (map[string]int64, error) {
	cases, err := junitxml.Parse(data)
	if err != nil {
		return nil, err
	}
	noPath := true
	got := junitxml.SumByID(cases, func(c junitxml.Case) string {
		if strings.Contains(c.Classname, "/") || strings.Contains(c.Classname, `\`) {
			noPath = false
			return filepath.ToSlash(c.Classname)
		}
		return c.Classname + " > " + c.Name
	})
	// jest-junit v16+ defaults classNameTemplate to the test title, not the
	// file path: every id would silently mismatch discovery. Fail loudly.
	if len(cases) > 0 && noPath {
		return nil, fmt.Errorf("no file-path classnames in JUnit XML; configure jest-junit with classNameTemplate: \"{filepath}\" (see docs/adapters/jest.md)")
	}
	return got, nil
}
