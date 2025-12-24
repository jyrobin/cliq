// Package guide provides man-page style documentation rendering for CLI applications.
//
// Guide content is loaded from YAML files (typically embedded with go:embed)
// and rendered to the terminal with sections, command references, and other
// documentation elements.
package guide

import (
	"io/fs"
)

// Index represents the top-level guide index, typically stored in index.yaml.
type Index struct {
	Title       string  `yaml:"title"`
	Description string  `yaml:"description"`
	Topics      []Topic `yaml:"topics"`
}

// Topic represents a single guide topic in the index.
type Topic struct {
	ID    string `yaml:"id"`    // Filename without .yaml extension
	Title string `yaml:"title"` // Full title
	Short string `yaml:"short"` // One-line description
}

// Content represents a single guide document.
type Content struct {
	Title       string     `yaml:"title"`
	Description string     `yaml:"description"`
	Sections    []Section  `yaml:"sections"`    // Free-form sections with content
	Categories  []Category `yaml:"categories"`  // Command categories (for command reference)
	Flags       []Flag     `yaml:"flags"`       // Global flags
	Environment []EnvVar   `yaml:"environment"` // Environment variables
}

// Section represents a named section with free-form content.
type Section struct {
	Name    string `yaml:"name"`
	Content string `yaml:"content"`
}

// Category groups related commands together.
type Category struct {
	Name     string    `yaml:"name"`
	Commands []Command `yaml:"commands"`
}

// Command represents a CLI command with usage and flags.
type Command struct {
	Name        string    `yaml:"name"`
	Usage       string    `yaml:"usage"`       // Usage pattern
	Short       string    `yaml:"short"`       // Short description
	Description string    `yaml:"description"` // Long description
	Flags       []CmdFlag `yaml:"flags"`       // Command-specific flags
}

// CmdFlag represents a command-specific flag.
type CmdFlag struct {
	Flag string `yaml:"flag"` // Flag name and args (e.g., "-m, --module <path>")
	Desc string `yaml:"desc"` // Flag description
}

// Flag represents a global flag.
type Flag struct {
	Flag        string `yaml:"flag"`
	Description string `yaml:"description"`
}

// EnvVar represents an environment variable.
type EnvVar struct {
	Var         string `yaml:"var"`
	Description string `yaml:"description"`
}

// Guide manages guide content loaded from an embedded filesystem.
type Guide struct {
	fs     fs.FS
	prefix string // Path prefix within fs (e.g., "guide")
}

// Options configures Guide behavior.
type Options struct {
	// Prefix is the directory prefix within the fs (default: "guide")
	Prefix string
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		Prefix: "guide",
	}
}
