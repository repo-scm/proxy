package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/repo-scm/proxy/config"
	"github.com/repo-scm/proxy/server"
)

const (
	shutdownTimeout = 5 * time.Second
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

	go func() {
		addr := parseAddress(serveAddress)
		fmt.Printf("Starting proxy server on %s\n", addr)
		fmt.Printf("Web UI available at %s/ui\n", addr)
		if err := httpServer.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				fmt.Printf("Server error: %v\n", err)
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	fmt.Println("Shutting down server...")

	_ = httpServer.Shutdown(shutdownCtx)

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
