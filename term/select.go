package term

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// SelectOptions configures select behavior.
type SelectOptions struct {
	// Title shown above the list
	Title string
	// Size is the number of visible items (default: 10)
	Size int
	// HideHelp hides the navigation hints
	HideHelp bool
	// StartIndex is the initial cursor position
	StartIndex int
	// Templates for rendering (optional)
	ActivePrefix   string // Default: "▸ "
	InactivePrefix string // Default: "  "
	// Filter enables type-to-filter mode
	Filter bool
}

// SelectResult contains the result of a select operation.
type SelectResult struct {
	Index     int    // Selected index (-1 if cancelled)
	Value     string // Selected value
	Cancelled bool   // True if user pressed Esc/Ctrl+C
}

// Select shows an interactive single-select menu.
// Returns the selected index and value. Supports ESC to cancel.
func Select(label string, items []string, opts SelectOptions) (*SelectResult, error) {
	if len(items) == 0 {
		return &SelectResult{Index: -1, Cancelled: true}, nil
	}

	if opts.Size == 0 {
		opts.Size = 10
	}
	if opts.ActivePrefix == "" {
		opts.ActivePrefix = "▸ "
	}
	if opts.InactivePrefix == "" {
		opts.InactivePrefix = "  "
	}

	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	defer term.Restore(fd, oldState)

	cursor := opts.StartIndex
	if cursor < 0 || cursor >= len(items) {
		cursor = 0
	}

	offset := 0 // Scroll offset for pagination
	filter := ""
	filteredItems := items
	filteredIndices := make([]int, len(items))
	for i := range items {
		filteredIndices[i] = i
	}

	// Calculate total lines needed
	totalLines := opts.Size + 1 // label + items
	if opts.Title != "" {
		totalLines++
	}
	if !opts.HideHelp {
		totalLines++
	}

	// Reserve space to prevent terminal scrolling issues
	for i := 0; i < totalLines; i++ {
		fmt.Println()
	}
	fmt.Printf("\033[%dA", totalLines)

	for {
		// Apply filter if enabled
		if opts.Filter && filter != "" {
			filteredItems = nil
			filteredIndices = nil
			lowerFilter := strings.ToLower(filter)
			for i, item := range items {
				if strings.Contains(strings.ToLower(item), lowerFilter) {
					filteredItems = append(filteredItems, item)
					filteredIndices = append(filteredIndices, i)
				}
			}
			if cursor >= len(filteredItems) {
				cursor = len(filteredItems) - 1
			}
			if cursor < 0 {
				cursor = 0
			}
		} else {
			filteredItems = items
			filteredIndices = make([]int, len(items))
			for i := range items {
				filteredIndices[i] = i
			}
		}

		// Adjust scroll offset
		if cursor < offset {
			offset = cursor
		}
		if cursor >= offset+opts.Size {
			offset = cursor - opts.Size + 1
		}

		// Render
		renderSelect(label, filteredItems, cursor, offset, opts, filter)

		// Read key
		key, char, err := readSelectKey()
		if err != nil {
			return nil, err
		}

		switch key {
		case selectKeyUp:
			if cursor > 0 {
				cursor--
			}
		case selectKeyDown:
			if cursor < len(filteredItems)-1 {
				cursor++
			}
		case selectKeyHome:
			cursor = 0
			offset = 0
		case selectKeyEnd:
			cursor = len(filteredItems) - 1
		case selectKeyPageUp:
			cursor -= opts.Size
			if cursor < 0 {
				cursor = 0
			}
		case selectKeyPageDown:
			cursor += opts.Size
			if cursor >= len(filteredItems) {
				cursor = len(filteredItems) - 1
			}
		case selectKeyEnter:
			if len(filteredItems) == 0 {
				continue
			}
			clearSelect(totalLines)
			originalIdx := filteredIndices[cursor]
			fmt.Printf("%s%s %s\n", opts.ActivePrefix, Cyan(label), Green(items[originalIdx]))
			return &SelectResult{
				Index: originalIdx,
				Value: items[originalIdx],
			}, nil
		case selectKeyEsc, selectKeyCtrlC:
			clearSelect(totalLines)
			return &SelectResult{
				Index:     -1,
				Cancelled: true,
			}, nil
		case selectKeyBackspace:
			if opts.Filter && len(filter) > 0 {
				filter = filter[:len(filter)-1]
			}
		case selectKeyChar:
			if opts.Filter && char >= 32 && char < 127 {
				filter += string(char)
			}
		}
	}
}

// SelectSimple is a convenience wrapper that returns just the selected value.
func SelectSimple(label string, items []string) (string, error) {
	result, err := Select(label, items, SelectOptions{})
	if err != nil {
		return "", err
	}
	if result.Cancelled {
		return "", fmt.Errorf("cancelled")
	}
	return result.Value, nil
}

// SelectIndex is a convenience wrapper that returns just the selected index.
func SelectIndex(label string, items []string) (int, error) {
	result, err := Select(label, items, SelectOptions{})
	if err != nil {
		return -1, err
	}
	if result.Cancelled {
		return -1, fmt.Errorf("cancelled")
	}
	return result.Index, nil
}

// Select key types
type selectKeyType int

const (
	selectKeyNone selectKeyType = iota
	selectKeyUp
	selectKeyDown
	selectKeyLeft
	selectKeyRight
	selectKeyHome
	selectKeyEnd
	selectKeyPageUp
	selectKeyPageDown
	selectKeyEnter
	selectKeyEsc
	selectKeyCtrlC
	selectKeyBackspace
	selectKeyChar
)

func readSelectKey() (selectKeyType, rune, error) {
	var buf [6]byte
	n, err := os.Stdin.Read(buf[:])
	if err != nil {
		return selectKeyNone, 0, err
	}
	if n == 0 {
		return selectKeyNone, 0, nil
	}

	// Handle escape sequences
	if buf[0] == 0x1b {
		if n == 1 {
			return selectKeyEsc, 0, nil
		}
		if buf[1] == '[' {
			switch buf[2] {
			case 'A':
				return selectKeyUp, 0, nil
			case 'B':
				return selectKeyDown, 0, nil
			case 'C':
				return selectKeyRight, 0, nil
			case 'D':
				return selectKeyLeft, 0, nil
			case 'H':
				return selectKeyHome, 0, nil
			case 'F':
				return selectKeyEnd, 0, nil
			case '5':
				if n > 3 && buf[3] == '~' {
					return selectKeyPageUp, 0, nil
				}
			case '6':
				if n > 3 && buf[3] == '~' {
					return selectKeyPageDown, 0, nil
				}
			}
		}
		return selectKeyEsc, 0, nil
	}

	// Control characters
	switch buf[0] {
	case 0x03: // Ctrl+C
		return selectKeyCtrlC, 0, nil
	case 0x0d, 0x0a: // Enter
		return selectKeyEnter, 0, nil
	case 0x7f, 0x08: // Backspace
		return selectKeyBackspace, 0, nil
	case 'k', 'K':
		return selectKeyUp, 0, nil
	case 'j', 'J':
		return selectKeyDown, 0, nil
	}

	// Regular character
	if buf[0] >= 32 && buf[0] < 127 {
		return selectKeyChar, rune(buf[0]), nil
	}

	return selectKeyNone, 0, nil
}

func renderSelect(label string, items []string, cursor, offset int, opts SelectOptions, filter string) {
	// Clear and render
	fmt.Print("\r\033[K")

	// Title
	if opts.Title != "" {
		fmt.Printf("%s\n\r\033[K", Bold(opts.Title))
	}

	// Label with optional filter
	if opts.Filter {
		if filter != "" {
			fmt.Printf("%s %s\n\r\033[K", Cyan(label), Dim("(filter: "+filter+")"))
		} else {
			fmt.Printf("%s %s\n\r\033[K", Cyan(label), Dim("(type to filter)"))
		}
	} else {
		fmt.Printf("%s\n\r\033[K", Cyan(label))
	}

	// Items
	visibleCount := opts.Size
	if visibleCount > len(items) {
		visibleCount = len(items)
	}

	for i := 0; i < opts.Size; i++ {
		idx := offset + i
		fmt.Print("\r\033[K")
		if idx < len(items) {
			prefix := opts.InactivePrefix
			text := items[idx]
			if idx == cursor {
				prefix = opts.ActivePrefix
				text = Bold(text)
			}
			fmt.Printf("%s%s\n", prefix, text)
		} else {
			fmt.Print("\n")
		}
	}

	// Help text
	fmt.Print("\r\033[K")
	if !opts.HideHelp {
		help := "↑↓: navigate, Enter: select, Esc: cancel"
		if opts.Filter {
			help = "↑↓: navigate, Enter: select, Esc: cancel, type: filter"
		}
		fmt.Print(Dim(help))
	}

	// Move cursor back to top
	// Lines printed: title (optional) + label + items (opts.Size)
	// Cursor is on help line (no newline after it)
	totalLines := opts.Size + 1 // label + items
	if opts.Title != "" {
		totalLines++
	}
	if !opts.HideHelp {
		totalLines++
	}
	fmt.Printf("\033[%dA\r", totalLines)
}

func clearSelect(lines int) {
	for i := 0; i < lines; i++ {
		fmt.Print("\r\033[K\n")
	}
	fmt.Printf("\033[%dA", lines)
}
