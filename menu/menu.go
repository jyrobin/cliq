package menu

import (
	"fmt"

	"github.com/jyrobin/cliq/term"
	"github.com/manifoldco/promptui"
	"gopkg.in/yaml.v3"
)

// Menu represents an interactive menu instance.
type Menu struct {
	config  *Config
	options Options
}

// New creates a new menu from a Config.
func New(cfg *Config, opts ...Options) *Menu {
	opt := DefaultOptions()
	if len(opts) > 0 {
		opt = opts[0]
	}
	return &Menu{
		config:  cfg,
		options: opt,
	}
}

// LoadConfig parses menu configuration from YAML data.
func LoadConfig(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse menu config: %w", err)
	}
	return &cfg, nil
}

// Run starts the interactive menu loop.
func (m *Menu) Run() error {
	return m.showMainMenu()
}

func (m *Menu) showMainMenu() error {
	for {
		// Build category list
		items := make([]string, len(m.config.Categories)+1)
		for i, cat := range m.config.Categories {
			items[i] = fmt.Sprintf("%s %s", cat.Icon, cat.Name)
		}
		items[len(m.config.Categories)] = "✕ Exit"

		prompt := promptui.Select{
			Label: m.config.Title,
			Items: items,
			Size:  m.options.Size,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}",
				Active:   "▸ {{ . | cyan }}",
				Inactive: "  {{ . }}",
				Selected: "✓ {{ . | green }}",
			},
		}

		idx, _, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				return nil
			}
			return err
		}

		if idx == len(m.config.Categories) {
			return nil // Exit
		}

		if err := m.showCategoryMenu(&m.config.Categories[idx]); err != nil {
			if err == promptui.ErrInterrupt {
				continue // Go back to main menu
			}
			return err
		}
	}
}

func (m *Menu) showCategoryMenu(cat *Category) error {
	for {
		// Build item list
		items := make([]string, len(cat.Items)+1)
		for i, item := range cat.Items {
			items[i] = fmt.Sprintf("%-30s %s", item.Name, term.Dim(item.Short))
		}
		items[len(cat.Items)] = "← Back"

		prompt := promptui.Select{
			Label: cat.Name,
			Items: items,
			Size:  m.options.Size + 2,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}",
				Active:   "▸ {{ . | cyan }}",
				Inactive: "  {{ . }}",
				Selected: "✓ {{ . | green }}",
			},
		}

		idx, _, err := prompt.Run()
		if err != nil {
			return err
		}

		if idx == len(cat.Items) {
			return nil // Back
		}

		if err := m.executeItem(&cat.Items[idx]); err != nil {
			if err == promptui.ErrInterrupt {
				continue
			}
			// Show error but continue
			fmt.Printf("\n%s\n\n", term.Error(err.Error()))
			continue
		}
	}
}

func (m *Menu) executeItem(item *Item) error {
	fmt.Println()

	switch item.Action {
	case "run":
		return m.runCommand(item)
	case "prompt":
		return m.promptAndRun(item)
	default:
		// Custom action - delegate to handler
		if m.options.ActionHandler != nil {
			return m.options.ActionHandler(item)
		}
		return fmt.Errorf("no handler for action: %s", item.Action)
	}
}
