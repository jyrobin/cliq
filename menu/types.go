// Package menu provides an interactive menu system for CLI applications.
package menu

// Config represents the menu configuration, typically loaded from YAML.
type Config struct {
	Title      string     `yaml:"title"`
	Categories []Category `yaml:"categories"`
}

// Category represents a group of related menu items.
type Category struct {
	Name  string `yaml:"name"`
	Icon  string `yaml:"icon"`
	Items []Item `yaml:"items"`
}

// Item represents a single menu action.
type Item struct {
	Name    string  `yaml:"name"`
	Short   string  `yaml:"short"`   // Short description
	Action  string  `yaml:"action"`  // Action type: run, prompt, custom
	Command string  `yaml:"command"` // Command template with {{.var}} placeholders
	Preview string  `yaml:"preview"` // Dry-run command variant
	Inputs  []Input `yaml:"inputs"`  // Input prompts for variables
	Output  string  `yaml:"output"`  // Output hint: "clipboard" for copy suggestion
	Topic   string  `yaml:"topic"`   // For custom actions (e.g., guide topic)
}

// Input represents a user input field.
type Input struct {
	ID      string `yaml:"id"`      // Variable name for substitution
	Label   string `yaml:"label"`   // Display label
	Hint    string `yaml:"hint"`    // Help text
	Default string `yaml:"default"` // Default value
}

// ActionHandler is called for custom actions (action != "run" or "prompt").
// Return nil to continue, error to show message and return to menu.
type ActionHandler func(item *Item) error

// Options configures menu behavior.
type Options struct {
	// Size is the number of visible items in the select menu (default: 10)
	Size int

	// ActionHandler handles custom actions (when Action is not "run" or "prompt")
	ActionHandler ActionHandler

	// BeforeRun is called before executing a command (for logging, confirmation, etc.)
	// Return false to cancel execution
	BeforeRun func(command string) bool

	// AfterRun is called after command execution
	AfterRun func(command string, err error)
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		Size: 10,
	}
}
