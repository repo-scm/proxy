package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/repo-scm/proxy/config"
	"github.com/repo-scm/proxy/server"
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
	srv := server.NewServer(cfg)
	httpServer := &http.Server{
		Addr:    serveAddress,
		Handler: srv.Handler(),
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	serverErr := make(chan error, 1)

	go func() {
		addr := parseAddress(serveAddress)
		fmt.Printf("Starting server on %s\n", addr)
		fmt.Printf("Web UI at %s/ui\n", addr)
		if err := httpServer.ListenAndServe(); err != nil {
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
	case <-quit:
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("Server error: %v\n", err)
		}
	}

	_ = httpServer.Shutdown(ctx)

	return nil
}

func parseAddress(addr string) string {
	switch {
	case addr == "" || addr[0] == ':':
		if addr == "" {
			return "http://localhost"
		}
		return fmt.Sprintf("http://localhost%s", addr)
	default:
		return fmt.Sprintf("http://%s", addr)
	}
}
