// Package adapter defines the extension point for test frameworks: one
// package + one Register call per framework. See CONTRIBUTING.md.
package adapter

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// Granularity is the unit a framework can shard at.
type Granularity int

const (
	GranularityTest Granularity = iota
	GranularityFile
)

// Test is one discoverable unit.
type Test struct {
	ID   string
	File string // may be empty when the framework has no file concept
}

// DiscoverConfig carries discovery options. Filter is runner-native filter
// syntax, passed through verbatim — shardfit never interprets it.
type DiscoverConfig struct {
	Filter string
	Input  string // test-list file path, or "-" for stdin; empty = native discovery
}

// Adapter is the extension point.
type Adapter interface {
	Name() string
	Discover(cfg DiscoverConfig) ([]Test, error)
	Granularity() Granularity
	ParseDurations(junitXML []byte) (map[string]int64, error)
}

var registry = map[string]Adapter{}

// Register adds an adapter under its Name.
func Register(a Adapter) { registry[a.Name()] = a }

// Get returns a registered adapter by name.
func Get(name string) (Adapter, error) {
	a, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown adapter %q (available: %s)", name, Names())
	}
	return a, nil
}

// All returns registered adapter names, sorted.
func All() []string {
	names := make([]string, 0, len(registry))
	for n := range registry {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Names returns registered adapter names as a joined list for messages.
func Names() string { return strings.Join(All(), ", ") }

// ReadList reads test ids from a file (or stdin for "-"), one per line.
// Blank lines are skipped.
func ReadList(ref string, stdin io.Reader) ([]Test, error) {
	var r io.Reader
	if ref == "-" {
		r = stdin
	} else {
		f, err := os.Open(ref)
		if err != nil {
			return nil, fmt.Errorf("open test list %s: %w", ref, err)
		}
		defer f.Close()
		r = f
	}
	var tests []Test
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		if id := strings.TrimSpace(sc.Text()); id != "" {
			tests = append(tests, Test{ID: id})
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read test list: %w", err)
	}
	return tests, nil
}
