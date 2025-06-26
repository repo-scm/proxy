package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/repo-scm/proxy/config"
	"github.com/repo-scm/proxy/monitor"
	"github.com/repo-scm/proxy/utils"
)

var (
	outputFile   string
	verboseQuery bool
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query available site",
	Run: func(cmd *cobra.Command, args []string) {
		var _path string
		ctx := context.Background()
		config := GetConfig()
		if outputFile != "" {
			_path = utils.ExpandTilde(outputFile)
			if _, err := os.Stat(_path); err == nil {
				_, _ = fmt.Fprintln(os.Stderr, "output file exists:", outputFile)
				os.Exit(1)
			}
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

	queryCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "output file")
	queryCmd.PersistentFlags().BoolVarP(&verboseQuery, "verbose", "v", false, "verbose mode")
}

func runQuery(ctx context.Context, cfg *config.Config, _path string) error {
	var buf string

	m := monitor.NewMonitor(cfg)

	site, err := m.GetAvailableSite()
	if err != nil {
		return err
	}

	if verboseQuery {
		s, err := json.Marshal(site)
		if err != nil {
			return err
		}
		buf = string(s)
	} else {
		buf = site.Host
	}

	if _path != "" {
		if err := os.WriteFile(_path, []byte(buf), utils.PermFile); err != nil {
			return err
		}
	} else {
		fmt.Println(buf)
	}

	return nil
}
