package main

import (
	"fmt"

	"github.com/jyrobin/cliq/term"
	"github.com/spf13/cobra"
)

var autocompleteCmd = &cobra.Command{
	Use:   "autocomplete",
	Short: "Demo autocomplete input",
	Long: `Demonstrates the autocomplete input feature.

Examples:
  cliq autocomplete              # Interactive demo with predefined options
  cliq autocomplete --fuzzy      # Fuzzy matching mode
  cliq autocomplete --path       # Path completion mode
  cliq autocomplete --command    # Command completion mode`,
	RunE: runAutocomplete,
}

var (
	fuzzyMode   bool
	pathMode    bool
	commandMode bool
)

func init() {
	autocompleteCmd.Flags().BoolVar(&fuzzyMode, "fuzzy", false, "Use fuzzy matching")
	autocompleteCmd.Flags().BoolVar(&pathMode, "path", false, "Path completion mode")
	autocompleteCmd.Flags().BoolVar(&commandMode, "command", false, "Command completion mode")
	rootCmd.AddCommand(autocompleteCmd)
}

func runAutocomplete(cmd *cobra.Command, args []string) error {
	if pathMode {
		return demoPathCompletion()
	}
	if commandMode {
		return demoCommandCompletion()
	}
	if fuzzyMode {
		return demoFuzzyCompletion()
	}
	return demoBasicCompletion()
}

func demoBasicCompletion() error {
	fmt.Println(term.Bold("Autocomplete Demo - Basic Mode"))
	fmt.Println(term.Dim("Type to filter, Tab to complete, ↑↓ to select, Enter to confirm"))
	fmt.Println()

	options := []string{
		"extension-create",
		"extension-delete",
		"extension-list",
		"extension-update",
		"trunk-create",
		"trunk-delete",
		"trunk-list",
		"route-inbound",
		"route-outbound",
		"config-apply",
		"config-export",
		"config-import",
	}

	result, err := term.Autocomplete(
		term.PrefixCompleter(options, false),
		term.AutocompleteOptions{
			Prompt:         "Command: ",
			Placeholder:    "start typing...",
			MaxSuggestions: 6,
		},
	)
	if err != nil {
		return err
	}

	if result.Cancelled {
		fmt.Println(term.Dim("Cancelled"))
		return nil
	}

	fmt.Printf("\nSelected: %s", term.Green(result.Value))
	if result.WasSuggested {
		fmt.Print(term.Dim(" (from suggestion)"))
	}
	fmt.Println()

	return nil
}

func demoFuzzyCompletion() error {
	fmt.Println(term.Bold("Autocomplete Demo - Fuzzy Mode"))
	fmt.Println(term.Dim("Characters match in order but not consecutively (e.g., 'ce' → CreateExtension)"))
	fmt.Println()

	options := []string{
		"CreateExtension",
		"DeleteExtension",
		"ListExtensions",
		"UpdateExtension",
		"CreateTrunk",
		"DeleteTrunk",
		"ListTrunks",
		"ApplyConfig",
		"ExportConfig",
		"ImportConfig",
	}

	result, err := term.Autocomplete(
		term.FuzzyCompleter(options),
		term.AutocompleteOptions{
			Prompt:         "Action: ",
			Placeholder:    "try 'ce' for CreateExtension...",
			MaxSuggestions: 6,
		},
	)
	if err != nil {
		return err
	}

	if result.Cancelled {
		fmt.Println(term.Dim("Cancelled"))
		return nil
	}

	fmt.Printf("\nSelected: %s\n", term.Green(result.Value))
	return nil
}

func demoPathCompletion() error {
	fmt.Println(term.Bold("Autocomplete Demo - Path Mode"))
	fmt.Println(term.Dim("Tab-complete filesystem paths"))
	fmt.Println()

	result, err := term.Autocomplete(
		term.PathCompleter(false),
		term.AutocompleteOptions{
			Prompt:         "Path: ",
			Placeholder:    "./",
			MaxSuggestions: 8,
			DefaultValue:   "./",
		},
	)
	if err != nil {
		return err
	}

	if result.Cancelled {
		fmt.Println(term.Dim("Cancelled"))
		return nil
	}

	fmt.Printf("\nSelected: %s\n", term.Green(result.Value))
	return nil
}

func demoCommandCompletion() error {
	fmt.Println(term.Bold("Autocomplete Demo - Command Mode"))
	fmt.Println(term.Dim("Command completion with argument suggestions"))
	fmt.Println()

	commands := []string{"get", "set", "list", "delete", "help"}
	argOptions := map[string][]string{
		"get":    {"extension", "trunk", "route", "config"},
		"set":    {"extension", "trunk", "route"},
		"list":   {"extensions", "trunks", "routes", "channels"},
		"delete": {"extension", "trunk", "route"},
		"help":   {"get", "set", "list", "delete"},
	}

	result, err := term.Autocomplete(
		term.CommandCompleter(commands, func(cmd, args string) []string {
			if opts, ok := argOptions[cmd]; ok {
				return term.PrefixCompleter(opts, false)(args)
			}
			return nil
		}),
		term.AutocompleteOptions{
			Prompt:         "$ ",
			Placeholder:    "command [args]",
			MaxSuggestions: 6,
		},
	)
	if err != nil {
		return err
	}

	if result.Cancelled {
		fmt.Println(term.Dim("Cancelled"))
		return nil
	}

	fmt.Printf("\nCommand: %s\n", term.Green(result.Value))
	return nil
}
