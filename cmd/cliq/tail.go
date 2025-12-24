package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/jyrobin/cliq/sh"
	"github.com/spf13/cobra"
)

var tailCmd = &cobra.Command{
	Use:   "tail <pattern>",
	Short: "Follow multiple files matching a glob pattern",
	Long: `Tail follows multiple files matching a glob pattern, merging their
output into a single stream with colored prefixes to distinguish sources.

Uses fsnotify for efficient file watching (no polling).

Examples:
  cliq tail "logs/*.log"
  cliq tail "services/*/app.log" --no-color
  cliq tail "/var/log/syslog" --from-start
  cliq tail "*.log" --prefix path`,
	Args: cobra.ExactArgs(1),
	RunE: runTail,
}

var (
	tailNoColor   bool
	tailFromStart bool
	tailPrefix    string
)

func init() {
	tailCmd.Flags().BoolVar(&tailNoColor, "no-color", false, "Disable colored output")
	tailCmd.Flags().BoolVar(&tailFromStart, "from-start", false, "Start from beginning of file")
	tailCmd.Flags().StringVar(&tailPrefix, "prefix", "name", "Prefix format: name, path, or none")
}

func runTail(cmd *cobra.Command, args []string) error {
	pattern := args[0]

	opts := sh.TailOptions{
		Colors:    sh.DefaultTailColors,
		Prefix:    tailPrefix,
		FromStart: tailFromStart,
	}

	if tailNoColor {
		opts.Colors = nil
	}

	// Handle Ctrl+C gracefully
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	return sh.Tail(ctx, pattern, opts)
}
