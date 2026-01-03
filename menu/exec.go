package menu

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/jyrobin/cliq/term"
)

func (m *Menu) runCommand(item *Item) error {
	return m.executeCommand(item.Command)
}

func (m *Menu) promptAndRun(item *Item) error {
	// Collect inputs
	values, err := m.collectInputs(item.Inputs)
	if err != nil {
		return err
	}

	// Substitute values in command
	command := SubstituteValues(item.Command, values)

	// Show preview if available
	if item.Preview != "" {
		preview := SubstituteValues(item.Preview, values)
		fmt.Printf("\nPreview:\n")
		fmt.Printf("  %s\n\n", preview)

		// Ask what to do
		options := []string{
			"Run preview (--dry)",
			"Run full command",
			"Copy command",
			"Cancel",
		}

		result, err := term.Select("What would you like to do?", options, term.SelectOptions{})
		if err != nil {
			return err
		}
		if result.Cancelled {
			return nil
		}

		switch result.Index {
		case 0: // Preview
			return m.executeCommand(preview)
		case 1: // Full
			return m.executeCommand(command)
		case 2: // Copy
			term.CopyableCommand(command)
			term.WaitForEnter()
			return nil
		case 3: // Cancel
			return nil
		}
	}

	// No preview - check if this is clipboard output
	if item.Output == "clipboard" {
		fmt.Printf("\nCommand:\n  %s\n\n", command)

		options := []string{
			"Run and show output",
			"Show command (copy manually)",
			"Cancel",
		}

		result, err := term.Select("This generates output to copy", options, term.SelectOptions{})
		if err != nil {
			return err
		}
		if result.Cancelled {
			return nil
		}

		switch result.Index {
		case 0:
			return m.executeCommand(command)
		case 1:
			term.CopyablePrompt(command)
			term.WaitForEnter()
			return nil
		case 2:
			return nil
		}
	}

	// Simple command - just run it
	return m.executeCommand(command)
}

func (m *Menu) collectInputs(inputs []Input) (map[string]string, error) {
	values := make(map[string]string)

	for _, input := range inputs {
		label := input.Label
		if input.Hint != "" {
			label = fmt.Sprintf("%s (%s)", label, input.Hint)
		}

		result, err := term.Autocomplete(nil, term.AutocompleteOptions{
			Prompt:       label + ": ",
			DefaultValue: input.Default,
		})
		if err != nil {
			return nil, err
		}
		if result.Cancelled {
			return nil, fmt.Errorf("cancelled")
		}

		if result.Value == "" && input.Default == "" {
			return nil, fmt.Errorf("input required: %s", input.Label)
		}

		values[input.ID] = result.Value
	}

	return values, nil
}

func (m *Menu) executeCommand(command string) error {
	// Call BeforeRun hook if set
	if m.options.BeforeRun != nil {
		if !m.options.BeforeRun(command) {
			return nil // Cancelled
		}
	}

	fmt.Printf("\nRunning: %s\n\n", command)

	err := RunCommand(command)

	// Call AfterRun hook if set
	if m.options.AfterRun != nil {
		m.options.AfterRun(command, err)
	}

	if err != nil {
		return err
	}

	fmt.Println()
	term.WaitForEnter()
	return nil
}

// SubstituteValues replaces {{.key}} placeholders with values from the map.
func SubstituteValues(template string, values map[string]string) string {
	result := template
	for k, v := range values {
		result = strings.ReplaceAll(result, "{{."+k+"}}", v)
	}
	return result
}

// RunCommand executes a shell command string.
func RunCommand(command string) error {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

// RunCommandCapture executes a command and returns its output.
func RunCommandCapture(command string) (string, error) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("command failed: %w", err)
	}

	return string(output), nil
}
