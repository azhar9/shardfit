// Package xunit adapts shardfit to xunit (.NET), reading JUnit-format XML
// produced by the JUnitTestLogger package.
package xunit

import (
	"fmt"
	"os"
	"strings"

	"github.com/azhar9/shardfit/internal/adapter"
	"github.com/azhar9/shardfit/internal/junitxml"
)

// Adapter implements adapter.Adapter for xunit.
type Adapter struct{}

// New returns a new Adapter.
func New() *Adapter { return &Adapter{} }

// Compile-time assertion: Adapter implements adapter.Adapter.
var _ adapter.Adapter = (*Adapter)(nil)

func (a *Adapter) Name() string { return "xunit" }

func (a *Adapter) Granularity() adapter.Granularity { return adapter.GranularityTest }

// Discover consumes a pre-generated test list (e.g. from
// `dotnet test --list-tests`).
func (a *Adapter) Discover(cfg adapter.DiscoverConfig) ([]adapter.Test, error) {
	if cfg.Input == "" {
		return nil, fmt.Errorf("xunit adapter requires --input (a test list file, or - for stdin)")
	}
	return adapter.ReadList(cfg.Input, os.Stdin)
}

// ParseDurations uses classname.method ids from the JUnitTestLogger output.
// JUnitTestLogger emits the full display name (namespace included) in the
// name attribute — exactly what `dotnet test --list-tests` prints — so the
// name is used as-is when it already carries the classname prefix.
func (a *Adapter) ParseDurations(data []byte) (map[string]int64, error) {
	cases, err := junitxml.Parse(data)
	if err != nil {
		return nil, err
	}
	return junitxml.SumByID(cases, func(c junitxml.Case) string {
		if strings.HasPrefix(c.Name, c.Classname+".") {
			return c.Name
		}
		return c.Classname + "." + c.Name
	}), nil
}
