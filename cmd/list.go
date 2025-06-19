package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/repo-scm/proxy/config"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available sites",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		config := GetConfig()
		if err := runList(ctx, config); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	},
}

// nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(ctx context.Context, cfg *config.Config) error {
	return nil
}
