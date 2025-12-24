package cobrautil

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// VersionCommand creates a simple version command.
func VersionCommand(version, commit, date string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\n", version)
			if commit != "" {
				fmt.Printf("Commit:  %s\n", commit)
			}
			if date != "" {
				fmt.Printf("Built:   %s\n", date)
			}
		},
	}
}

// Execute runs a root command and exits with appropriate code.
func Execute(cmd *cobra.Command) {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// AddVerboseFlag adds a --verbose/-v flag to a command.
func AddVerboseFlag(cmd *cobra.Command, verbose *bool) {
	cmd.PersistentFlags().BoolVarP(verbose, "verbose", "v", false, "verbose output")
}

// AddDryRunFlag adds a --dry-run flag to a command.
func AddDryRunFlag(cmd *cobra.Command, dryRun *bool) {
	cmd.PersistentFlags().BoolVar(dryRun, "dry-run", false, "show what would be done without making changes")
}

// AddOutputFlag adds a --output/-o flag for output format.
func AddOutputFlag(cmd *cobra.Command, output *string, defaultFormat string) {
	cmd.PersistentFlags().StringVarP(output, "output", "o", defaultFormat, "output format (json, yaml, table)")
}

// MustMarkRequired marks a flag as required and panics on error.
func MustMarkRequired(cmd *cobra.Command, name string) {
	if err := cmd.MarkFlagRequired(name); err != nil {
		panic(fmt.Sprintf("failed to mark flag %s as required: %v", name, err))
	}
}

// RunFunc is a simplified RunE that doesn't need args.
type RunFunc func() error

// SimpleRunE wraps a RunFunc as a cobra.Command.RunE.
func SimpleRunE(fn RunFunc) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return fn()
	}
}

// ArgsRunE wraps a function that takes args as a cobra.Command.RunE.
func ArgsRunE(fn func(args []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return fn(args)
	}
}
