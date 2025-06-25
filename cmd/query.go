package cmd

import (
	"context"
	"fmt"
	"github.com/repo-scm/proxy/utils"
	"os"

	"github.com/spf13/cobra"

	"github.com/repo-scm/proxy/config"
)

var (
	outputFile   string
	verboseQuery bool
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query available site",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		config := GetConfig()
		_path := utils.ExpandTilde(outputFile)
		if _, err := os.Stat(_path); err == nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		if err := runQuery(ctx, config, _path); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	},
}

// nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(queryCmd)

	queryCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "output.json", "output file")
	queryCmd.PersistentFlags().BoolVarP(&verboseQuery, "verbose", "v", false, "verbose mode")
}

func runQuery(ctx context.Context, cfg *config.Config, _path string) error {
	return nil
}
