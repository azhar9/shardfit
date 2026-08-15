// Package generic adapts any runner that can produce a test list and JUnit
// XML — the zero-code path for frameworks without a native adapter.
package generic

import (
	"fmt"
	"os"

	"github.com/azhar9/shardfit/internal/adapter"
	"github.com/azhar9/shardfit/internal/junitxml"
)

// Adapter implements adapter.Adapter for arbitrary test ids.
type Adapter struct{}

// New returns a new Adapter.
func New() *Adapter { return &Adapter{} }

// Compile-time assertion: Adapter implements adapter.Adapter.
var _ adapter.Adapter = (*Adapter)(nil)

func (a *Adapter) Name() string { return "generic" }

func (a *Adapter) Granularity() adapter.Granularity { return adapter.GranularityTest }

// Discover consumes a user-supplied test list; generic has no native runner.
func (a *Adapter) Discover(cfg adapter.DiscoverConfig) ([]adapter.Test, error) {
	if cfg.Input == "" {
		return nil, fmt.Errorf("generic adapter requires --input (a test list file, or - for stdin)")
	}
	return adapter.ReadList(cfg.Input, os.Stdin)
}

// ParseDurations uses "classname.name" ids, or the bare name when classname
// is empty. Retries (duplicate cases) are summed by SumByID.
func (a *Adapter) ParseDurations(data []byte) (map[string]int64, error) {
	cases, err := junitxml.Parse(data)
	if err != nil {
		return nil, err
	}
	return junitxml.SumByID(cases, func(c junitxml.Case) string {
		if c.Classname == "" {
			return c.Name
		}
		return c.Classname + "." + c.Name
	}), nil
}
