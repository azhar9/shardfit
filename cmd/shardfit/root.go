package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/azhar9/shardfit/internal/adapter"
)

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "shardfit",
		Short:         "Split test suites into duration-balanced buckets",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	// one command tree per registered adapter: shardfit <adapter> <verb>
	for _, name := range adapter.All() {
		a, err := adapter.Get(name)
		if err != nil {
			panic(err) // registration bug
		}
		ac := &cobra.Command{Use: name, Short: fmt.Sprintf("%s adapter", name)}
		ac.AddCommand(newDiscoverCmd(a), newSplitCmd(a), newReportCmd(a))
		root.AddCommand(ac)
	}
	return root
}
