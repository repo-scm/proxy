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
}

func runList(ctx context.Context, cfg *config.Config) error {
	if err := listTable(ctx, cfg.Gerrits); err != nil {
		return err
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
