// Package term provides terminal utilities with zero external dependencies.
package term

// ANSI escape codes for styling
const (
	reset     = "\033[0m"
	bold      = "\033[1m"
	dim       = "\033[2m"
	italic    = "\033[3m"
	underline = "\033[4m"

	// Colors
	black   = "\033[30m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	white   = "\033[37m"

	// Bright colors
	brightBlack   = "\033[90m"
	brightRed     = "\033[91m"
	brightGreen   = "\033[92m"
	brightYellow  = "\033[93m"
	brightBlue    = "\033[94m"
	brightMagenta = "\033[95m"
	brightCyan    = "\033[96m"
	brightWhite   = "\033[97m"
)

// Style applies ANSI styling to text
func Style(s string, codes ...string) string {
	if len(codes) == 0 {
		return s
	}
	prefix := ""
	for _, c := range codes {
		prefix += c
	}
	return prefix + s + reset
}

// Dim returns dimmed text
func Dim(s string) string {
	return dim + s + reset
}

// Bold returns bold text
func Bold(s string) string {
	return bold + s + reset
}

// Italic returns italic text
func Italic(s string) string {
	return italic + s + reset
}

// Underline returns underlined text
func Underline(s string) string {
	return underline + s + reset
}

// Color functions

func Black(s string) string   { return black + s + reset }
func Red(s string) string     { return red + s + reset }
func Green(s string) string   { return green + s + reset }
func Yellow(s string) string  { return yellow + s + reset }
func Blue(s string) string    { return blue + s + reset }
func Magenta(s string) string { return magenta + s + reset }
func Cyan(s string) string    { return cyan + s + reset }
func White(s string) string   { return white + s + reset }

// Bright color functions

func BrightBlack(s string) string   { return brightBlack + s + reset }
func BrightRed(s string) string     { return brightRed + s + reset }
func BrightGreen(s string) string   { return brightGreen + s + reset }
func BrightYellow(s string) string  { return brightYellow + s + reset }
func BrightBlue(s string) string    { return brightBlue + s + reset }
func BrightMagenta(s string) string { return brightMagenta + s + reset }
func BrightCyan(s string) string    { return brightCyan + s + reset }
func BrightWhite(s string) string   { return brightWhite + s + reset }

// Success returns green text with checkmark
func Success(s string) string {
	return green + "✓ " + s + reset
}

// Error returns red text with X
func Error(s string) string {
	return red + "✗ " + s + reset
}

// Warning returns yellow text with warning sign
func Warning(s string) string {
	return yellow + "⚠ " + s + reset
}

// Info returns cyan text with info sign
func Info(s string) string {
	return cyan + "ℹ " + s + reset
}
