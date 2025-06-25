package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/repo-scm/proxy/config"
	"github.com/repo-scm/proxy/utils"
)

var (
	verboseList bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all sites",
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

	listCmd.PersistentFlags().BoolVarP(&verboseList, "verbose", "v", false, "verbose mode")
}

func runList(ctx context.Context, cfg *config.Config) error {
	if verboseList {
		return listTable(ctx, cfg.Gerrits)
	}

	for key, val := range cfg.Gerrits {
		fmt.Printf("NAME:%s, LOCATION:%s, WEIGHT:%.1f, HTTP:%s\n", key, val.Location, val.Weight, val.Http.Url)
	}

	return nil
}

func listTable(ctx context.Context, sites map[string]config.Gerrit) error {
	data := [][]string{
		{"NAME", "LOCATION", "WEIGHT", "HTTP", "SSH"},
	}

	for key, val := range sites {
		data = append(data, []string{key, val.Location, fmt.Sprintf("%.1f", val.Weight), val.Http.Url, fmt.Sprintf("ssh://%s:%d", val.Ssh.Host, val.Ssh.Port)})
	}

	if err := utils.WriteTable(ctx, data); err != nil {
		return errors.Wrap(err, "failed to write table\n")
	}

	return nil
}
