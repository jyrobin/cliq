// Package cobrautil provides Cobra integration utilities.
package cobrautil

import (
	"fmt"

	"github.com/jyrobin/cliq/menu"
	"github.com/spf13/cobra"
)

// MenuCommand creates a Cobra command that runs an interactive menu.
func MenuCommand(cfg *menu.Config, opts ...menu.Options) *cobra.Command {
	return &cobra.Command{
		Use:     "menu",
		Aliases: []string{"i", "interactive"},
		Short:   "Interactive menu",
		Long:    fmt.Sprintf("Launch interactive menu: %s", cfg.Title),
		RunE: func(cmd *cobra.Command, args []string) error {
			m := menu.New(cfg, opts...)
			return m.Run()
		},
	}
}

// MenuCommandWithLoader creates a Cobra command that loads menu config lazily.
// This is useful when the config comes from an embedded filesystem.
func MenuCommandWithLoader(loader func() (*menu.Config, error), opts ...menu.Options) *cobra.Command {
	return &cobra.Command{
		Use:     "menu",
		Aliases: []string{"i", "interactive"},
		Short:   "Interactive menu",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loader()
			if err != nil {
				return fmt.Errorf("failed to load menu config: %w", err)
			}
			m := menu.New(cfg, opts...)
			return m.Run()
		},
	}
}

// MenuConfig holds configuration for building a menu command.
type MenuConfig struct {
	// Use is the command name (default: "menu")
	Use string
	// Aliases are alternative command names (default: ["i", "interactive"])
	Aliases []string
	// Short is the short description
	Short string
	// Long is the long description
	Long string
}

// DefaultMenuConfig returns sensible defaults for menu command configuration.
func DefaultMenuConfig() MenuConfig {
	return MenuConfig{
		Use:     "menu",
		Aliases: []string{"i", "interactive"},
		Short:   "Interactive menu",
		Long:    "Launch an interactive menu for common tasks.",
	}
}

// MenuCommandCustom creates a Cobra command with custom configuration.
func MenuCommandCustom(cfg *menu.Config, cmdCfg MenuConfig, opts ...menu.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     cmdCfg.Use,
		Aliases: cmdCfg.Aliases,
		Short:   cmdCfg.Short,
		Long:    cmdCfg.Long,
		RunE: func(cmd *cobra.Command, args []string) error {
			m := menu.New(cfg, opts...)
			return m.Run()
		},
	}
	return cmd
}
