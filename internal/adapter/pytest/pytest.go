// Package pytest adapts shardfit to pytest.
package pytest

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

// Adapter implements adapter.Adapter for pytest.
type Adapter struct{}

// New returns a new Adapter.
func New() *Adapter { return &Adapter{} }

// Compile-time assertion: Adapter implements adapter.Adapter.
var _ adapter.Adapter = (*Adapter)(nil)

func (a *Adapter) Name() string { return "pytest" }

func (a *Adapter) Granularity() adapter.Granularity { return adapter.GranularityTest }

// Discover runs `pytest --collect-only -q`, passing Filter through verbatim,
// and parses the test ids. An Input list bypasses pytest entirely.
func (a *Adapter) Discover(cfg adapter.DiscoverConfig) ([]adapter.Test, error) {
	if cfg.Input != "" {
		tests, err := adapter.ReadList(cfg.Input, os.Stdin)
		if err != nil {
			return nil, err
		}
		for i := range tests {
			if j := strings.Index(tests[i].ID, "::"); j >= 0 {
				tests[i].File = tests[i].ID[:j]
			}
		}
		return tests, nil
	}
	args := []string{"--collect-only", "-q"}
	if cfg.Filter != "" {
		args = append(adapter.SplitFilter(cfg.Filter), args...)
	}
	out, err := exec.Command("pytest", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("pytest --collect-only failed: %w (output: %s)", err, out)
	}
	var tests []adapter.Test
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		id := strings.TrimSpace(sc.Text())
		if id == "" || !strings.Contains(id, "::") {
			continue
		}
		file := id
		if i := strings.Index(file, "::"); i >= 0 {
			file = file[:i]
		}
		tests = append(tests, adapter.Test{ID: id, File: file})
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read pytest output: %w", err)
	}
	return tests, nil
}

// ParseDurations maps JUnit XML testcases to pytest collect ids
// (path::Class::test). pytest's junitxml classname is a dotted module path,
// so the file part is reconstructed against the working tree; when the file
// is gone, ids degrade to a slashed-path form derived from the classname.
func (a *Adapter) ParseDurations(data []byte) (map[string]int64, error) {
	cases, err := junitxml.Parse(data)
	if err != nil {
		return nil, err
	}
	return junitxml.SumByID(cases, func(c junitxml.Case) string {
		file, class := fileFor(c.Classname)
		if class == "" {
			return file + "::" + c.Name
		}
		return file + "::" + class + "::" + c.Name
	}), nil
}

// fileFor maps "tests.test_api.TestClass" to ("tests/test_api.py",
// "TestClass") by trying candidate paths against the filesystem, then
// package __init__.py candidates, falling back to a slashed-path form
// derived from the dotted classname.
func fileFor(classname string) (file, class string) {
	parts := strings.Split(classname, ".")
	for n := len(parts); n >= 1; n-- {
		cand := strings.Join(parts[:n], "/") + ".py"
		if _, err := os.Stat(filepath.FromSlash(cand)); err == nil {
			return cand, strings.Join(parts[n:], "::")
		}
	}
	for n := len(parts); n >= 1; n-- {
		cand := strings.Join(parts[:n], "/") + "/__init__.py"
		if _, err := os.Stat(filepath.FromSlash(cand)); err == nil {
			return cand, strings.Join(parts[n:], "::")
		}
	}
	return strings.ReplaceAll(classname, ".", "/") + ".py", ""
}
