package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/repo-scm/proxy/config"
)

var (
	BuildTime string
	CommitID  string
)

var (
	cfgFile string
	cfgData *config.Config
)

var rootCmd = &cobra.Command{
	Use:     "proxy",
	Short:   "git sites proxy",
	Version: BuildTime + "-" + CommitID,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

// nolint:gochecknoinits
func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default $HOME/.repo-scm/proxy.yaml)")

	rootCmd.Root().CompletionOptions.DisableDefaultCmd = true
}

func initConfig() {
	var err error

	if cfgData, err = config.LoadConfig(cfgFile); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func GetConfig() *config.Config {
	return cfgData
}
