package term

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// PrefixCompleter returns a Completer that matches options by prefix.
func PrefixCompleter(options []string, caseSensitive bool) Completer {
	return func(input string) []string {
		if input == "" {
			return options
		}

		var matches []string
		for _, opt := range options {
			if caseSensitive {
				if strings.HasPrefix(opt, input) {
					matches = append(matches, opt)
				}
			} else {
				if strings.HasPrefix(strings.ToLower(opt), strings.ToLower(input)) {
					matches = append(matches, opt)
				}
			}
		}
		return matches
	}
}

// ContainsCompleter returns a Completer that matches options containing the input.
func ContainsCompleter(options []string, caseSensitive bool) Completer {
	return func(input string) []string {
		if input == "" {
			return options
		}

		var matches []string
		for _, opt := range options {
			if caseSensitive {
				if strings.Contains(opt, input) {
					matches = append(matches, opt)
				}
			} else {
				if strings.Contains(strings.ToLower(opt), strings.ToLower(input)) {
					matches = append(matches, opt)
				}
			}
		}
		return matches
	}
}

// FuzzyCompleter returns a Completer that matches options using fuzzy matching.
// Characters must appear in order but not consecutively (e.g., "ce" matches "CreateExtension").
func FuzzyCompleter(options []string) Completer {
	return func(input string) []string {
		if input == "" {
			return options
		}

		type scored struct {
			option string
			score  int
		}

		var matches []scored
		inputLower := strings.ToLower(input)

		for _, opt := range options {
			optLower := strings.ToLower(opt)
			if score := fuzzyScore(inputLower, optLower); score > 0 {
				matches = append(matches, scored{opt, score})
			}
		}

		// Sort by score (higher is better)
		sort.Slice(matches, func(i, j int) bool {
			return matches[i].score > matches[j].score
		})

		result := make([]string, len(matches))
		for i, m := range matches {
			result[i] = m.option
		}
		return result
	}
}

// fuzzyScore returns a score for how well pattern matches text.
// Returns 0 if no match. Higher score = better match.
func fuzzyScore(pattern, text string) int {
	if len(pattern) == 0 {
		return 1
	}
	if len(pattern) > len(text) {
		return 0
	}

	score := 0
	patternIdx := 0
	lastMatchIdx := -1
	consecutive := 0

	for i := 0; i < len(text) && patternIdx < len(pattern); i++ {
		if text[i] == pattern[patternIdx] {
			patternIdx++
			score += 10

			// Bonus for consecutive matches
			if lastMatchIdx == i-1 {
				consecutive++
				score += consecutive * 5
			} else {
				consecutive = 0
			}

			// Bonus for matching at word boundaries
			if i == 0 || text[i-1] == ' ' || text[i-1] == '_' || text[i-1] == '-' ||
				(i > 0 && isLower(text[i-1]) && isUpper(text[i])) {
				score += 20
			}

			lastMatchIdx = i
		}
	}

	// All pattern characters must be matched
	if patternIdx < len(pattern) {
		return 0
	}

	return score
}

func isLower(b byte) bool {
	return b >= 'a' && b <= 'z'
}

func isUpper(b byte) bool {
	return b >= 'A' && b <= 'Z'
}

// PathCompleter returns a Completer for filesystem paths.
// It expands ~ and completes directories and files.
func PathCompleter(dirsOnly bool) Completer {
	return func(input string) []string {
		// Expand ~
		if strings.HasPrefix(input, "~") {
			home, err := os.UserHomeDir()
			if err == nil {
				input = home + input[1:]
			}
		}

		// Determine directory and prefix to match
		dir := "."
		prefix := input

		if input == "" {
			dir = "."
			prefix = ""
		} else if strings.HasSuffix(input, "/") {
			dir = input
			prefix = ""
		} else {
			dir = filepath.Dir(input)
			prefix = filepath.Base(input)
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil
		}

		var matches []string
		for _, entry := range entries {
			name := entry.Name()

			// Skip hidden files unless prefix starts with .
			if strings.HasPrefix(name, ".") && !strings.HasPrefix(prefix, ".") {
				continue
			}

			// Match prefix
			if prefix != "" && !strings.HasPrefix(strings.ToLower(name), strings.ToLower(prefix)) {
				continue
			}

			// Filter dirs only if requested
			if dirsOnly && !entry.IsDir() {
				continue
			}

			// Build full path
			fullPath := filepath.Join(dir, name)
			if entry.IsDir() {
				fullPath += "/"
			}

			// Clean up the path for display
			if strings.HasPrefix(input, "./") {
				fullPath = "./" + strings.TrimPrefix(fullPath, dir+"/")
				if !strings.HasPrefix(fullPath, "./") {
					fullPath = "./" + fullPath
				}
			}

			matches = append(matches, fullPath)
		}

		sort.Strings(matches)
		return matches
	}
}

// ChainCompleters combines multiple completers, returning results from all.
func ChainCompleters(completers ...Completer) Completer {
	return func(input string) []string {
		seen := make(map[string]bool)
		var results []string

		for _, c := range completers {
			for _, r := range c(input) {
				if !seen[r] {
					seen[r] = true
					results = append(results, r)
				}
			}
		}

		return results
	}
}

// MapCompleter returns a Completer that transforms suggestions.
func MapCompleter(completer Completer, transform func(string) string) Completer {
	return func(input string) []string {
		results := completer(input)
		for i, r := range results {
			results[i] = transform(r)
		}
		return results
	}
}

// FilterCompleter returns a Completer that filters suggestions.
func FilterCompleter(completer Completer, filter func(string) bool) Completer {
	return func(input string) []string {
		results := completer(input)
		filtered := results[:0]
		for _, r := range results {
			if filter(r) {
				filtered = append(filtered, r)
			}
		}
		return filtered
	}
}

// StaticCompleter returns a Completer with static suggestions (no filtering).
func StaticCompleter(options ...string) Completer {
	return func(input string) []string {
		return options
	}
}

// CommandCompleter creates a completer for command-style input.
// It completes the first word from commands, then uses argCompleter for arguments.
func CommandCompleter(commands []string, argCompleter func(cmd, args string) []string) Completer {
	return func(input string) []string {
		parts := strings.SplitN(input, " ", 2)

		// If no space yet, complete commands
		if len(parts) == 1 {
			return PrefixCompleter(commands, false)(input)
		}

		// Otherwise, use arg completer
		if argCompleter != nil {
			cmd := parts[0]
			args := parts[1]
			suggestions := argCompleter(cmd, args)

			// Prepend command to each suggestion
			for i, s := range suggestions {
				suggestions[i] = cmd + " " + s
			}
			return suggestions
		}

		return nil
	}
}
