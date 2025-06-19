package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/repo-scm/proxy/config"
)

var (
	serveAddress string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run proxy server",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		config := GetConfig()
		if err := runServe(ctx, config); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	},
}

// nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.PersistentFlags().StringVarP(&serveAddress, "address", "a", ":9090", "serve address")
}

func runServe(ctx context.Context, cfg *config.Config) error {
	return nil
}
