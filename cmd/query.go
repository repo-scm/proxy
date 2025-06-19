package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/repo-scm/proxy/config"
)

var (
	outputFile  string
	verboseMode bool
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query available site",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		config := GetConfig()
		if err := runQuery(ctx, config); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	},
}

// nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(queryCmd)

	queryCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "output.json", "output file")
	queryCmd.PersistentFlags().BoolVarP(&verboseMode, "verbose", "v", false, "verbose mode")
}

func runQuery(ctx context.Context, cfg *config.Config) error {
	return nil
}
