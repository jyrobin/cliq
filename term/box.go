package term

import (
	"fmt"
	"strings"
)

// Box styles
type BoxStyle struct {
	TopLeft     string
	TopRight    string
	BottomLeft  string
	BottomRight string
	Horizontal  string
	Vertical    string
}

var (
	// BoxSingle uses single-line box drawing characters
	BoxSingle = BoxStyle{
		TopLeft:     "┌",
		TopRight:    "┐",
		BottomLeft:  "└",
		BottomRight: "┘",
		Horizontal:  "─",
		Vertical:    "│",
	}

	// BoxDouble uses double-line box drawing characters
	BoxDouble = BoxStyle{
		TopLeft:     "╔",
		TopRight:    "╗",
		BottomLeft:  "╚",
		BottomRight: "╝",
		Horizontal:  "═",
		Vertical:    "║",
	}

	// BoxRounded uses rounded corners
	BoxRounded = BoxStyle{
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "╰",
		BottomRight: "╯",
		Horizontal:  "─",
		Vertical:    "│",
	}

	// BoxASCII uses ASCII characters for compatibility
	BoxASCII = BoxStyle{
		TopLeft:     "+",
		TopRight:    "+",
		BottomLeft:  "+",
		BottomRight: "+",
		Horizontal:  "-",
		Vertical:    "|",
	}
)

// Box draws a box around content
func Box(content string, width int, style BoxStyle) string {
	if width < 4 {
		width = 60
	}

	innerWidth := width - 2
	lines := strings.Split(content, "\n")

	var result strings.Builder

	// Top border
	result.WriteString(style.TopLeft)
	result.WriteString(strings.Repeat(style.Horizontal, innerWidth))
	result.WriteString(style.TopRight)
	result.WriteString("\n")

	// Content lines
	for _, line := range lines {
		result.WriteString(style.Vertical)
		result.WriteString(" ")
		if len(line) > innerWidth-2 {
			line = line[:innerWidth-2]
		}
		result.WriteString(line)
		result.WriteString(strings.Repeat(" ", innerWidth-len(line)-1))
		result.WriteString(style.Vertical)
		result.WriteString("\n")
	}

	// Bottom border
	result.WriteString(style.BottomLeft)
	result.WriteString(strings.Repeat(style.Horizontal, innerWidth))
	result.WriteString(style.BottomRight)

	return result.String()
}

// Copyable displays content in a box with copy instructions
func Copyable(content string, hint string) {
	fmt.Println()
	fmt.Println(Box(content, 65, BoxSingle))
	fmt.Println()
	if hint != "" {
		fmt.Println(Dim(hint))
		fmt.Println()
	}
}

// CopyableCommand displays a command to be copied
func CopyableCommand(command string) {
	Copyable(command, "Copy and run this command.")
}

// CopyablePrompt displays a command that generates output to copy
func CopyablePrompt(command string) {
	Copyable(command, "Run this command and copy the output to Claude.")
}
