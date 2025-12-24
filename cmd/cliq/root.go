package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cliq",
	Short: "CLI toolkit utilities",
	Long: `cliq provides command-line utilities for working with
files, HTTP endpoints, WebSockets, and data transformation.

Useful for scripting, testing, and ad-hoc debugging.`,
}

func init() {
	rootCmd.AddCommand(tailCmd)
}
