// Package junit adapts shardfit to JUnit (Java, surefire reports).
package junit

import (
	"fmt"
	"os"

	"github.com/azhar9/shardfit/internal/adapter"
	"github.com/azhar9/shardfit/internal/junitxml"
)

// Adapter implements adapter.Adapter for JUnit.
type Adapter struct{}

// New returns a new Adapter.
func New() *Adapter { return &Adapter{} }

// Compile-time assertion: Adapter implements adapter.Adapter.
var _ adapter.Adapter = (*Adapter)(nil)

func (a *Adapter) Name() string { return "junit" }

func (a *Adapter) Granularity() adapter.Granularity { return adapter.GranularityTest }

// Discover consumes a pre-generated test list: Java discovery needs a
// compiled build, which shardfit does not drive.
func (a *Adapter) Discover(cfg adapter.DiscoverConfig) ([]adapter.Test, error) {
	if cfg.Input == "" {
		return nil, fmt.Errorf("junit adapter requires --input (a test list file, e.g. from a surefire scan, or - for stdin)")
	}
	return adapter.ReadList(cfg.Input, os.Stdin)
}

// ParseDurations uses surefire's classname.method ids; retries are summed.
func (a *Adapter) ParseDurations(data []byte) (map[string]int64, error) {
	cases, err := junitxml.Parse(data)
	if err != nil {
		return nil, err
	}
	return junitxml.SumByID(cases, func(c junitxml.Case) string {
		return c.Classname + "." + c.Name
	}), nil
}
