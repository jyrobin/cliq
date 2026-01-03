package term

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"golang.org/x/term"
)

// Completer provides suggestions based on current input.
type Completer func(input string) []string

// AutocompleteOptions configures autocomplete behavior.
type AutocompleteOptions struct {
	// Prompt is the text shown before the input (default: "> ")
	Prompt string
	// Placeholder is shown when input is empty (dimmed)
	Placeholder string
	// MaxSuggestions limits displayed suggestions (default: 5)
	MaxSuggestions int
	// CaseSensitive controls matching (default: false)
	CaseSensitive bool
	// DefaultValue pre-fills the input
	DefaultValue string
	// Validate checks input on enter (return error to reject)
	Validate func(input string) error
}

// AutocompleteResult contains the result of autocomplete input.
type AutocompleteResult struct {
	Value       string // The entered/selected value
	WasSuggested bool   // True if value came from suggestions
	Cancelled   bool   // True if user pressed Esc/Ctrl+C
}

// Autocomplete shows an input with tab-completion suggestions.
// Press Tab to autocomplete, Up/Down to navigate suggestions, Enter to confirm.
func Autocomplete(completer Completer, opts AutocompleteOptions) (*AutocompleteResult, error) {
	if opts.Prompt == "" {
		opts.Prompt = "> "
	}
	if opts.MaxSuggestions == 0 {
		opts.MaxSuggestions = 5
	}

	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	defer term.Restore(fd, oldState)

	input := []rune(opts.DefaultValue)
	cursor := len(input)
	suggestions := []string{}
	selectedIdx := -1 // -1 means no suggestion selected
	errorMsg := ""

	// Initial suggestions
	if completer != nil {
		suggestions = completer(string(input))
	}

	for {
		// Render
		renderAutocomplete(opts.Prompt, input, cursor, suggestions, selectedIdx, opts.MaxSuggestions, opts.Placeholder, errorMsg)
		errorMsg = ""

		// Read key
		key, err := readAutocompleteKey()
		if err != nil {
			return nil, err
		}

		switch key.Type {
		case keyTypeChar:
			// Insert character
			input = insertRune(input, cursor, key.Char)
			cursor++
			selectedIdx = -1
			if completer != nil {
				suggestions = completer(string(input))
			}

		case keyTypeBackspace:
			if cursor > 0 {
				input = deleteRune(input, cursor-1)
				cursor--
				selectedIdx = -1
				if completer != nil {
					suggestions = completer(string(input))
				}
			}

		case keyTypeDelete:
			if cursor < len(input) {
				input = deleteRune(input, cursor)
				selectedIdx = -1
				if completer != nil {
					suggestions = completer(string(input))
				}
			}

		case keyTypeLeft:
			if cursor > 0 {
				cursor--
			}

		case keyTypeRight:
			if cursor < len(input) {
				cursor++
			}

		case keyTypeHome:
			cursor = 0

		case keyTypeEnd:
			cursor = len(input)

		case keyTypeUp:
			if len(suggestions) > 0 {
				if selectedIdx < 0 {
					selectedIdx = 0
				} else if selectedIdx > 0 {
					selectedIdx--
				}
			}

		case keyTypeDown:
			if len(suggestions) > 0 {
				if selectedIdx < len(suggestions)-1 {
					selectedIdx++
				}
			}

		case keyTypeTab:
			// Autocomplete with first/selected suggestion
			if len(suggestions) > 0 {
				idx := selectedIdx
				if idx < 0 {
					idx = 0
				}
				input = []rune(suggestions[idx])
				cursor = len(input)
				selectedIdx = -1
				if completer != nil {
					suggestions = completer(string(input))
				}
			}

		case keyTypeEnter:
			value := string(input)
			if selectedIdx >= 0 && selectedIdx < len(suggestions) {
				value = suggestions[selectedIdx]
			}

			// Validate if provided
			if opts.Validate != nil {
				if err := opts.Validate(value); err != nil {
					errorMsg = err.Error()
					continue
				}
			}

			clearAutocomplete(opts.MaxSuggestions + 2)
			fmt.Printf("%s%s\n", opts.Prompt, value)
			return &AutocompleteResult{
				Value:        value,
				WasSuggested: selectedIdx >= 0,
			}, nil

		case keyTypeEscape, keyTypeCtrlC:
			clearAutocomplete(opts.MaxSuggestions + 2)
			return &AutocompleteResult{Cancelled: true}, nil

		case keyTypeCtrlU:
			// Clear line
			input = []rune{}
			cursor = 0
			selectedIdx = -1
			if completer != nil {
				suggestions = completer("")
			}

		case keyTypeCtrlW:
			// Delete word backward
			if cursor > 0 {
				// Find start of word
				start := cursor - 1
				for start > 0 && input[start-1] != ' ' {
					start--
				}
				input = append(input[:start], input[cursor:]...)
				cursor = start
				selectedIdx = -1
				if completer != nil {
					suggestions = completer(string(input))
				}
			}
		}
	}
}

// AutocompleteSimple is a convenience wrapper with a static list of options.
func AutocompleteSimple(prompt string, options []string) (string, error) {
	result, err := Autocomplete(func(input string) []string {
		if input == "" {
			return options
		}
		lower := strings.ToLower(input)
		var matches []string
		for _, opt := range options {
			if strings.Contains(strings.ToLower(opt), lower) {
				matches = append(matches, opt)
			}
		}
		return matches
	}, AutocompleteOptions{Prompt: prompt})

	if err != nil {
		return "", err
	}
	if result.Cancelled {
		return "", fmt.Errorf("cancelled")
	}
	return result.Value, nil
}

// Key types for autocomplete
type keyType int

const (
	keyTypeChar keyType = iota
	keyTypeBackspace
	keyTypeDelete
	keyTypeLeft
	keyTypeRight
	keyTypeUp
	keyTypeDown
	keyTypeHome
	keyTypeEnd
	keyTypeTab
	keyTypeEnter
	keyTypeEscape
	keyTypeCtrlC
	keyTypeCtrlU
	keyTypeCtrlW
)

type autocompleteKey struct {
	Type keyType
	Char rune
}

func readAutocompleteKey() (autocompleteKey, error) {
	var buf [6]byte
	n, err := os.Stdin.Read(buf[:])
	if err != nil {
		return autocompleteKey{}, err
	}
	if n == 0 {
		return autocompleteKey{}, nil
	}

	// Handle escape sequences
	if buf[0] == 0x1b {
		if n == 1 {
			return autocompleteKey{Type: keyTypeEscape}, nil
		}
		if buf[1] == '[' {
			switch buf[2] {
			case 'A':
				return autocompleteKey{Type: keyTypeUp}, nil
			case 'B':
				return autocompleteKey{Type: keyTypeDown}, nil
			case 'C':
				return autocompleteKey{Type: keyTypeRight}, nil
			case 'D':
				return autocompleteKey{Type: keyTypeLeft}, nil
			case 'H':
				return autocompleteKey{Type: keyTypeHome}, nil
			case 'F':
				return autocompleteKey{Type: keyTypeEnd}, nil
			case '3':
				if n > 3 && buf[3] == '~' {
					return autocompleteKey{Type: keyTypeDelete}, nil
				}
			}
		}
		return autocompleteKey{Type: keyTypeEscape}, nil
	}

	// Control characters
	switch buf[0] {
	case 0x03: // Ctrl+C
		return autocompleteKey{Type: keyTypeCtrlC}, nil
	case 0x09: // Tab
		return autocompleteKey{Type: keyTypeTab}, nil
	case 0x0d, 0x0a: // Enter
		return autocompleteKey{Type: keyTypeEnter}, nil
	case 0x15: // Ctrl+U
		return autocompleteKey{Type: keyTypeCtrlU}, nil
	case 0x17: // Ctrl+W
		return autocompleteKey{Type: keyTypeCtrlW}, nil
	case 0x7f, 0x08: // Backspace
		return autocompleteKey{Type: keyTypeBackspace}, nil
	}

	// Regular character (handle UTF-8)
	r, _ := utf8.DecodeRune(buf[:n])
	if r != utf8.RuneError && r >= 32 {
		return autocompleteKey{Type: keyTypeChar, Char: r}, nil
	}

	return autocompleteKey{}, nil
}

func renderAutocomplete(prompt string, input []rune, cursor int, suggestions []string, selectedIdx, maxSuggestions int, placeholder, errorMsg string) {
	// Move to start of area and clear
	fmt.Print("\r\033[K")

	// Show error if any
	if errorMsg != "" {
		fmt.Printf("%s\n\r\033[K", Red(errorMsg))
	}

	// Show prompt and input
	fmt.Print(prompt)
	if len(input) == 0 && placeholder != "" {
		fmt.Print(Dim(placeholder))
		// Move cursor back to after prompt
		fmt.Printf("\r\033[%dC", len(prompt))
	} else {
		fmt.Print(string(input))
		// Position cursor
		if cursor < len(input) {
			fmt.Printf("\033[%dD", len(input)-cursor)
		}
	}

	// Save cursor position
	fmt.Print("\033[s")

	// Show suggestions below
	displayCount := len(suggestions)
	if displayCount > maxSuggestions {
		displayCount = maxSuggestions
	}

	for i := 0; i < maxSuggestions; i++ {
		fmt.Print("\n\r\033[K")
		if i < displayCount {
			prefix := "  "
			text := suggestions[i]
			if i == selectedIdx {
				prefix = Cyan("> ")
				text = Bold(text)
			} else {
				text = Dim(text)
			}
			fmt.Printf("%s%s", prefix, text)
		}
	}

	// Show hint
	fmt.Print("\n\r\033[K")
	if len(suggestions) > maxSuggestions {
		fmt.Print(Dim(fmt.Sprintf("  ... and %d more", len(suggestions)-maxSuggestions)))
	} else if len(suggestions) > 0 {
		fmt.Print(Dim("  Tab: complete, ↑↓: select"))
	}

	// Restore cursor position
	fmt.Print("\033[u")
}

func clearAutocomplete(lines int) {
	// Move down and clear each line, then move back up
	for i := 0; i < lines; i++ {
		fmt.Print("\n\r\033[K")
	}
	// Move back up
	fmt.Printf("\033[%dA\r", lines)
}

func insertRune(runes []rune, pos int, r rune) []rune {
	result := make([]rune, len(runes)+1)
	copy(result[:pos], runes[:pos])
	result[pos] = r
	copy(result[pos+1:], runes[pos:])
	return result
}

func deleteRune(runes []rune, pos int) []rune {
	result := make([]rune, len(runes)-1)
	copy(result[:pos], runes[:pos])
	copy(result[pos:], runes[pos+1:])
	return result
}
