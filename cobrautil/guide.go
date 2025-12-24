package cobrautil

import (
	"fmt"
	"io/fs"

	"github.com/jyrobin/cliq/guide"
	"github.com/spf13/cobra"
)

// GuideCommand creates a Cobra command for displaying guides.
// The fsys should contain guide YAML files in the specified prefix directory.
func GuideCommand(fsys fs.FS, opts guide.Options) *cobra.Command {
	g := guide.New(fsys, opts)

	return &cobra.Command{
		Use:   "guide [topic]",
		Short: "Show usage guide",
		Long: `Display comprehensive usage guides.

Without arguments, lists available topics.
With a topic argument, shows the detailed guide for that topic.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return g.ShowIndex()
			}
			return g.ShowTopic(args[0])
		},
	}
}

// GuideConfig holds configuration for building a guide command.
type GuideConfig struct {
	// Use is the command name (default: "guide")
	Use string
	// Aliases are alternative command names
	Aliases []string
	// Short is the short description
	Short string
	// Long is the long description
	Long string
	// UsageHint is appended to error messages (e.g., "Run 'mycli guide' to see available topics")
	UsageHint string
}

// DefaultGuideConfig returns sensible defaults for guide command configuration.
func DefaultGuideConfig() GuideConfig {
	return GuideConfig{
		Use:   "guide",
		Short: "Show usage guide",
		Long: `Display comprehensive usage guides.

Without arguments, lists available topics.
With a topic argument, shows the detailed guide for that topic.`,
	}
}

// GuideCommandCustom creates a Cobra command with custom configuration.
func GuideCommandCustom(fsys fs.FS, opts guide.Options, cmdCfg GuideConfig) *cobra.Command {
	g := guide.New(fsys, opts)

	usageHint := cmdCfg.UsageHint
	if usageHint == "" {
		usageHint = fmt.Sprintf("Run '%s' to see available topics", cmdCfg.Use)
	}

	cmd := &cobra.Command{
		Use:     cmdCfg.Use + " [topic]",
		Aliases: cmdCfg.Aliases,
		Short:   cmdCfg.Short,
		Long:    cmdCfg.Long,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return g.ShowIndex()
			}
			if !g.HasTopic(args[0]) {
				return fmt.Errorf("unknown topic: %s\n%s", args[0], usageHint)
			}
			return g.ShowTopic(args[0])
		},
	}
	return cmd
}

// GuideHandler returns a function suitable for use as menu.ActionHandler
// when menu items have action: "guide" and topic: "<topic-id>".
func GuideHandler(fsys fs.FS, opts guide.Options) func(topic string) error {
	g := guide.New(fsys, opts)
	return func(topic string) error {
		return g.ShowTopic(topic)
	}
}
