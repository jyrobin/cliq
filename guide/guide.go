package guide

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// New creates a Guide from an embedded filesystem.
// The opts.Prefix specifies the directory within fs containing guide files.
func New(fsys fs.FS, opts Options) *Guide {
	if opts.Prefix == "" {
		opts.Prefix = "guide"
	}
	return &Guide{
		fs:     fsys,
		prefix: opts.Prefix,
	}
}

// ShowIndex displays the list of available topics.
func (g *Guide) ShowIndex() error {
	return g.ShowIndexTo(os.Stdout)
}

// ShowIndexTo writes the topic list to the given writer.
func (g *Guide) ShowIndexTo(w io.Writer) error {
	index, err := g.LoadIndex()
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "=== %s ===\n\n", index.Title)
	if index.Description != "" {
		fmt.Fprintf(w, "%s\n", strings.TrimSpace(index.Description))
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w, "Available topics:")
	fmt.Fprintln(w)

	for _, topic := range index.Topics {
		fmt.Fprintf(w, "  %-12s %s\n", topic.ID, topic.Short)
	}

	fmt.Fprintln(w)
	return nil
}

// ShowTopic displays a specific guide topic.
func (g *Guide) ShowTopic(topic string) error {
	return g.ShowTopicTo(os.Stdout, topic)
}

// ShowTopicTo writes a specific guide topic to the given writer.
func (g *Guide) ShowTopicTo(w io.Writer, topic string) error {
	content, err := g.LoadTopic(topic)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "=== %s ===\n\n", content.Title)

	if content.Description != "" {
		fmt.Fprintf(w, "%s\n", strings.TrimSpace(content.Description))
		fmt.Fprintln(w)
	}

	// Render sections (for concepts, workflow)
	for _, section := range content.Sections {
		fmt.Fprintf(w, "## %s\n\n", section.Name)
		fmt.Fprintf(w, "%s\n", strings.TrimSpace(section.Content))
		fmt.Fprintln(w)
	}

	// Render categories with commands (for commands.yaml)
	for _, cat := range content.Categories {
		fmt.Fprintf(w, "## %s\n\n", cat.Name)
		for _, cmd := range cat.Commands {
			fmt.Fprintf(w, "  %s\n", cmd.Name)
			fmt.Fprintf(w, "    %s\n", cmd.Usage)
			if cmd.Short != "" {
				fmt.Fprintf(w, "    %s\n", cmd.Short)
			}
			for _, f := range cmd.Flags {
				fmt.Fprintf(w, "      %s  %s\n", f.Flag, f.Desc)
			}
			fmt.Fprintln(w)
		}
	}

	// Render flags (for flags.yaml)
	for _, flag := range content.Flags {
		fmt.Fprintf(w, "## %s\n\n", flag.Flag)
		fmt.Fprintf(w, "%s\n", strings.TrimSpace(flag.Description))
		fmt.Fprintln(w)
	}

	// Render environment variables
	if len(content.Environment) > 0 {
		fmt.Fprintln(w, "## Environment Variables")
		fmt.Fprintln(w)
		for _, env := range content.Environment {
			fmt.Fprintf(w, "  %s\n\n", env.Var)
			fmt.Fprintf(w, "%s\n", strings.TrimSpace(env.Description))
			fmt.Fprintln(w)
		}
	}

	return nil
}

// LoadIndex loads and parses the guide index.
func (g *Guide) LoadIndex() (*Index, error) {
	data, err := fs.ReadFile(g.fs, g.prefix+"/index.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read guide index: %w", err)
	}

	var index Index
	if err := yaml.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse guide index: %w", err)
	}

	return &index, nil
}

// LoadTopic loads and parses a specific guide topic.
func (g *Guide) LoadTopic(topic string) (*Content, error) {
	filename := fmt.Sprintf("%s/%s.yaml", g.prefix, topic)
	data, err := fs.ReadFile(g.fs, filename)
	if err != nil {
		return nil, fmt.Errorf("unknown topic: %s", topic)
	}

	var content Content
	if err := yaml.Unmarshal(data, &content); err != nil {
		return nil, fmt.Errorf("failed to parse guide %s: %w", topic, err)
	}

	return &content, nil
}

// Topics returns the list of available topics.
func (g *Guide) Topics() ([]Topic, error) {
	index, err := g.LoadIndex()
	if err != nil {
		return nil, err
	}
	return index.Topics, nil
}

// HasTopic checks if a topic exists.
func (g *Guide) HasTopic(topic string) bool {
	filename := fmt.Sprintf("%s/%s.yaml", g.prefix, topic)
	_, err := fs.ReadFile(g.fs, filename)
	return err == nil
}
