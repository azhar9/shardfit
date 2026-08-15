// Command shardfit splits test suites into duration-balanced buckets.
package main

import (
	"fmt"
	"os"

	"github.com/azhar9/shardfit/internal/adapter"
	"github.com/azhar9/shardfit/internal/adapter/generic"
	"github.com/azhar9/shardfit/internal/adapter/jest"
	adapterjunit "github.com/azhar9/shardfit/internal/adapter/junit"
	"github.com/azhar9/shardfit/internal/adapter/pytest"
	"github.com/azhar9/shardfit/internal/adapter/xunit"
)

func init() {
	adapter.Register(pytest.New())
	adapter.Register(jest.New())
	adapter.Register(adapterjunit.New())
	adapter.Register(xunit.New())
	adapter.Register(generic.New())
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "shardfit:", err)
		os.Exit(1)
	}
}
