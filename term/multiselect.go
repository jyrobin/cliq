package term

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// SelectItem represents an item in a multi-select list
type SelectItem struct {
	ID          string // Unique identifier
	Label       string // Display label
	Description string // Optional description
	Required    bool   // If true, cannot be deselected
	Selected    bool   // Current selection state
}

// MultiSelectOptions configures the multi-select behavior
type MultiSelectOptions struct {
	Title       string // Title shown above the list
	HelpText    string // Help text shown below (default: navigation hints)
	AllSelected bool   // If true, all items start selected (default: false)
}

// MultiSelect shows an interactive multi-select and returns selected item IDs
func MultiSelect(items []SelectItem, opts MultiSelectOptions) ([]string, error) {
	if len(items) == 0 {
		return nil, nil
	}

	// Apply default selection
	if opts.AllSelected {
		for i := range items {
			items[i].Selected = true
		}
	}

	// Ensure required items are selected
	for i := range items {
		if items[i].Required {
			items[i].Selected = true
		}
	}

	cursor := 0
	for {
		// Clear screen and draw
		fmt.Print("\033[H\033[2J")

		title := opts.Title
		if title == "" {
			title = "Select items"
		}
		fmt.Printf("%s %s\n\n", Bold(title), Dim("(space=toggle, enter=done, q=cancel)"))

		for i, item := range items {
			prefix := "  "
			if i == cursor {
				prefix = "> "
			}

			checkbox := "[ ]"
			if item.Selected {
				checkbox = "[x]"
			}

			label := item.Label
			if i == cursor {
				label = Bold(label)
			}

			line := fmt.Sprintf("%s%s %s", prefix, checkbox, label)
			if item.Description != "" {
				line += " " + Dim(item.Description)
			}
			if item.Required {
				line += Dim(" *")
			}
			fmt.Println(line)
		}

		if opts.HelpText != "" {
			fmt.Printf("\n%s\n", Dim(opts.HelpText))
		} else {
			fmt.Printf("\n%s\n", Dim("* = required (cannot deselect)"))
		}

		// Read key
		key, err := readKey()
		if err != nil {
			return nil, err
		}

		switch key {
		case 'k', 'K', keyUp:
			if cursor > 0 {
				cursor--
			}
		case 'j', 'J', keyDown:
			if cursor < len(items)-1 {
				cursor++
			}
		case ' ':
			// Toggle selection
			if items[cursor].Required && items[cursor].Selected {
				fmt.Print("\a") // Bell - can't deselect required
			} else {
				items[cursor].Selected = !items[cursor].Selected
			}
		case 'a', 'A':
			// Select all
			for i := range items {
				items[i].Selected = true
			}
		case 'n', 'N':
			// Deselect all (except required)
			for i := range items {
				if !items[i].Required {
					items[i].Selected = false
				}
			}
		case keyEnter, '\n':
			// Done - return selected IDs
			fmt.Print("\033[H\033[2J")
			var selected []string
			for _, item := range items {
				if item.Selected {
					selected = append(selected, item.ID)
				}
			}
			return selected, nil
		case 'q', 'Q', keyEsc:
			fmt.Print("\033[H\033[2J")
			return nil, fmt.Errorf("cancelled")
		}
	}
}

// MultiSelectDeselected is like MultiSelect but returns deselected item IDs
// Useful when you want to know what was skipped rather than what was selected
func MultiSelectDeselected(items []SelectItem, opts MultiSelectOptions) ([]string, error) {
	if len(items) == 0 {
		return nil, nil
	}

	// Apply default selection
	if opts.AllSelected {
		for i := range items {
			items[i].Selected = true
		}
	}

	// Ensure required items are selected
	for i := range items {
		if items[i].Required {
			items[i].Selected = true
		}
	}

	cursor := 0
	for {
		// Clear screen and draw
		fmt.Print("\033[H\033[2J")

		title := opts.Title
		if title == "" {
			title = "Select items"
		}
		fmt.Printf("%s %s\n\n", Bold(title), Dim("(space=toggle, enter=done, q=cancel)"))

		for i, item := range items {
			prefix := "  "
			if i == cursor {
				prefix = "> "
			}

			checkbox := "[ ]"
			if item.Selected {
				checkbox = "[x]"
			}

			label := item.Label
			if i == cursor {
				label = Bold(label)
			}

			line := fmt.Sprintf("%s%s %s", prefix, checkbox, label)
			if item.Description != "" {
				line += " " + Dim(item.Description)
			}
			if item.Required {
				line += Dim(" *")
			}
			fmt.Println(line)
		}

		if opts.HelpText != "" {
			fmt.Printf("\n%s\n", Dim(opts.HelpText))
		} else {
			fmt.Printf("\n%s\n", Dim("* = required (cannot deselect)"))
		}

		// Read key
		key, err := readKey()
		if err != nil {
			return nil, err
		}

		switch key {
		case 'k', 'K', keyUp:
			if cursor > 0 {
				cursor--
			}
		case 'j', 'J', keyDown:
			if cursor < len(items)-1 {
				cursor++
			}
		case ' ':
			if items[cursor].Required && items[cursor].Selected {
				fmt.Print("\a")
			} else {
				items[cursor].Selected = !items[cursor].Selected
			}
		case 'a', 'A':
			for i := range items {
				items[i].Selected = true
			}
		case 'n', 'N':
			for i := range items {
				if !items[i].Required {
					items[i].Selected = false
				}
			}
		case keyEnter, '\n':
			fmt.Print("\033[H\033[2J")
			var deselected []string
			for _, item := range items {
				if !item.Selected {
					deselected = append(deselected, item.ID)
				}
			}
			return deselected, nil
		case 'q', 'Q', keyEsc:
			fmt.Print("\033[H\033[2J")
			return nil, fmt.Errorf("cancelled")
		}
	}
}

// Key codes
const (
	keyUp    = 0x1b5b41
	keyDown  = 0x1b5b42
	keyEnter = 0x0d
	keyEsc   = 0x1b
)

// readKey reads a single keypress in raw mode
func readKey() (int, error) {
	fd := int(os.Stdin.Fd())

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return 0, err
	}
	defer term.Restore(fd, oldState)

	var buf [3]byte
	n, err := os.Stdin.Read(buf[:])
	if err != nil {
		return 0, err
	}

	if n == 0 {
		return 0, nil
	}

	// Handle escape sequences (arrow keys)
	if buf[0] == 0x1b && n >= 3 {
		return int(buf[0])<<16 | int(buf[1])<<8 | int(buf[2]), nil
	}

	return int(buf[0]), nil
}
