package term

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// WaitForEnter waits for the user to press Enter
func WaitForEnter() {
	fmt.Print("Press Enter to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

// WaitForEnterWithMessage waits for Enter with a custom message
func WaitForEnterWithMessage(msg string) {
	fmt.Print(msg)
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

// Confirm asks a yes/no question, returns true for yes
// Default is used when user presses Enter without input
func Confirm(prompt string, defaultYes bool) bool {
	suffix := " [y/N]: "
	if defaultYes {
		suffix = " [Y/n]: "
	}

	fmt.Print(prompt + suffix)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return defaultYes
	}

	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return defaultYes
	}

	return input == "y" || input == "yes"
}

// ReadLine reads a single line of input with an optional default
func ReadLine(prompt string, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return defaultValue
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	return input
}

// ReadPassword reads input without echoing (basic implementation)
// For production use, consider golang.org/x/term
func ReadPassword(prompt string) string {
	fmt.Print(prompt)
	// Note: This doesn't actually hide input - would need x/term for that
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}
