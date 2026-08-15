package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/azhar9/shardfit/internal/adapter"
)

func newDiscoverCmd(a adapter.Adapter) *cobra.Command {
	var filter, input string
	cmd := &cobra.Command{
		Use:   "discover",
		Short: "List test ids, one per line",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			tests, err := a.Discover(adapter.DiscoverConfig{Filter: filter, Input: input})
			if err != nil {
				return err
			}
			for _, t := range tests {
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), t.ID); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&filter, "filter", "", "runner-native filter, forwarded verbatim (e.g. -k unit)")
	cmd.Flags().StringVar(&input, "input", "", "test list file instead of native discovery (- for stdin)")
	return cmd
}
