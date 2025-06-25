package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
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

	queryCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "output file (.json|.txt)")
	queryCmd.PersistentFlags().BoolVarP(&verboseQuery, "verbose", "v", false, "verbose mode")
}

func runQuery(ctx context.Context, cfg *config.Config, _path string) error {
	var buf string
	var err error

	m := monitor.NewMonitor(cfg)

	site, err := m.GetAvailableSite()
	if err != nil {
		return err
	}

	table := [][]string{
		{"NAME", "URL", "HOST", "HEALTHY", "RESPONSETIME", "CONNECTIONS", "QUEUESIZE", "SCORE", "LASTCHECK", "ERROR"},
	}

	if verboseQuery {
		var b []string
		table = append(table, []string{site.Name, site.Url, site.Host, strconv.FormatBool(site.Healthy), fmt.Sprintf("%dms", site.ResponseTime), strconv.Itoa(site.Connections), strconv.Itoa(site.QueueSize), strconv.Itoa(site.Score), site.LastCheck.Format(time.RFC3339), site.Error})
		for index, item := range table[0] {
			b = append(b, fmt.Sprintf("%s:%s", item, table[1][index]))
		}
		buf = strings.Join(b, ", ")
	} else {
		buf = fmt.Sprintf("NAME:%s, URL:%s, CONNECTIONS:%d, QUEUESIZE:%d", site.Name, site.Url, site.Connections, site.QueueSize)
	}

	if _path != "" {
		return queryOutput(_path, buf)
	} else {
		if verboseQuery {
			if err = queryTable(ctx, table); err != nil {
				return err
			}
		} else {
			fmt.Println(buf)
		}
	}

	return nil
}

func queryTable(ctx context.Context, data [][]string) error {
	if err := utils.WriteTable(ctx, data); err != nil {
		return errors.Wrap(err, "failed to write table\n")
	}

	return nil
}

func queryOutput(name string, data string) error {
	var buf string
	var err error

	ext := filepath.Ext(name)
	switch ext {
	case ".json":
		buf, err = convertToJson(data)
		if err != nil {
			return errors.Wrap(err, "failed to convert to json\n")
		}
	case ".txt":
		buf = data
	default:
		return errors.New("invalid file extension\n")
	}

	if err := os.WriteFile(name, []byte(buf), utils.PermFile); err != nil {
		return errors.Wrap(err, "failed to write file\n")
	}

	return nil
}

func convertToJson(data string) (string, error) {
	pairs := strings.Split(data, ", ")
	buf := make(map[string]string)

	for _, pair := range pairs {
		kv := strings.SplitN(pair, ":", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			buf[key] = value
		}
	}

	jsonBytes, err := json.MarshalIndent(buf, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
