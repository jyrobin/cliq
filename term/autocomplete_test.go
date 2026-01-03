package term

import (
	"strings"
	"testing"
)

func TestInsertRune(t *testing.T) {
	tests := []struct {
		input    []rune
		pos      int
		r        rune
		expected string
	}{
		{[]rune("hello"), 0, 'X', "Xhello"},
		{[]rune("hello"), 5, 'X', "helloX"},
		{[]rune("hello"), 2, 'X', "heXllo"},
		{[]rune(""), 0, 'X', "X"},
	}

	for _, tt := range tests {
		result := insertRune(tt.input, tt.pos, tt.r)
		if string(result) != tt.expected {
			t.Errorf("insertRune(%q, %d, %q) = %q, want %q",
				string(tt.input), tt.pos, tt.r, string(result), tt.expected)
		}
	}
}

func TestDeleteRune(t *testing.T) {
	tests := []struct {
		input    []rune
		pos      int
		expected string
	}{
		{[]rune("hello"), 0, "ello"},
		{[]rune("hello"), 4, "hell"},
		{[]rune("hello"), 2, "helo"},
		{[]rune("X"), 0, ""},
	}

	for _, tt := range tests {
		result := deleteRune(tt.input, tt.pos)
		if string(result) != tt.expected {
			t.Errorf("deleteRune(%q, %d) = %q, want %q",
				string(tt.input), tt.pos, string(result), tt.expected)
		}
	}
}

func TestPrefixCompleter(t *testing.T) {
	options := []string{"apple", "application", "banana", "bandana"}
	completer := PrefixCompleter(options, false)

	tests := []struct {
		input    string
		expected []string
	}{
		{"", options},
		{"a", []string{"apple", "application"}},
		{"app", []string{"apple", "application"}},
		{"b", []string{"banana", "bandana"}},
		{"ban", []string{"banana", "bandana"}},
		{"bana", []string{"banana"}},
		{"xyz", []string{}},
		{"A", []string{"apple", "application"}}, // case insensitive
	}

	for _, tt := range tests {
		result := completer(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("PrefixCompleter(%q) returned %d items, want %d",
				tt.input, len(result), len(tt.expected))
			continue
		}
		for i, r := range result {
			if r != tt.expected[i] {
				t.Errorf("PrefixCompleter(%q)[%d] = %q, want %q",
					tt.input, i, r, tt.expected[i])
			}
		}
	}
}

func TestContainsCompleter(t *testing.T) {
	options := []string{"apple", "pineapple", "banana"}
	completer := ContainsCompleter(options, false)

	tests := []struct {
		input    string
		expected []string
	}{
		{"", options},
		{"apple", []string{"apple", "pineapple"}},
		{"pine", []string{"pineapple"}},
		{"an", []string{"banana"}},
	}

	for _, tt := range tests {
		result := completer(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("ContainsCompleter(%q) returned %d items, want %d",
				tt.input, len(result), len(tt.expected))
		}
	}
}

func TestFuzzyCompleter(t *testing.T) {
	options := []string{"CreateExtension", "DeleteExtension", "ListExtensions"}
	completer := FuzzyCompleter(options)

	tests := []struct {
		input       string
		shouldMatch []string
	}{
		{"ce", []string{"CreateExtension"}},
		{"de", []string{"DeleteExtension"}},
		{"ext", []string{"CreateExtension", "DeleteExtension", "ListExtensions"}},
		{"le", []string{"DeleteExtension", "ListExtensions"}},
	}

	for _, tt := range tests {
		result := completer(tt.input)
		for _, expected := range tt.shouldMatch {
			found := false
			for _, r := range result {
				if r == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("FuzzyCompleter(%q) should contain %q, got %v",
					tt.input, expected, result)
			}
		}
	}
}

func TestPathCompleterMock(t *testing.T) {
	// Test the path splitting logic
	tests := []struct {
		input string
		dir   string
		base  string
	}{
		{"/home/user/", "/home/user/", ""},
		{"/home/user/doc", "/home/user/", "doc"},
		{"./test", "./", "test"},
		{"test", "", "test"},
	}

	for _, tt := range tests {
		lastSlash := strings.LastIndex(tt.input, "/")
		var dir, base string
		if lastSlash >= 0 {
			dir = tt.input[:lastSlash+1]
			base = tt.input[lastSlash+1:]
		} else {
			dir = ""
			base = tt.input
		}

		if dir != tt.dir {
			t.Errorf("path %q: dir = %q, want %q", tt.input, dir, tt.dir)
		}
		if base != tt.base {
			t.Errorf("path %q: base = %q, want %q", tt.input, base, tt.base)
		}
	}
}
