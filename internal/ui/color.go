package ui

import (
	"fmt"
	"os"
)

var noColor bool

func init() {
	// Respect NO_COLOR env (https://no-color.org/)
	if os.Getenv("NO_COLOR") != "" {
		noColor = true
	}
	// Disable color if not a terminal
	fi, err := os.Stdout.Stat()
	if err != nil || fi.Mode()&os.ModeCharDevice == 0 {
		noColor = true
	}
}

const (
	reset   = "\033[0m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	cyan    = "\033[36m"
	bold    = "\033[1m"
	dim     = "\033[2m"
)

func colorize(color, msg string) string {
	if noColor {
		return msg
	}
	return color + msg + reset
}

// Success prints a green success message.
func Success(msg string) { fmt.Println(colorize(green, "✓ "+msg)) }

// Info prints a blue info message.
func Info(msg string) { fmt.Println(colorize(cyan, msg)) }

// Warn prints a yellow warning message.
func Warn(msg string) { fmt.Fprintln(os.Stderr, colorize(yellow, "⚠ "+msg)) }

// Error prints a red error message.
func Error(msg string) { fmt.Fprintln(os.Stderr, colorize(red, "✗ "+msg)) }

// Step prints a bold step message (for progress).
func Step(msg string) { fmt.Println(colorize(bold, msg)) }

// Dim prints a dimmed message (for secondary info).
func Dim(msg string) { fmt.Println(colorize(dim, msg)) }

// Successf formats and prints a green success message.
func Successf(format string, a ...any) { Success(fmt.Sprintf(format, a...)) }

// Infof formats and prints a blue info message.
func Infof(format string, a ...any) { Info(fmt.Sprintf(format, a...)) }

// Warnf formats and prints a yellow warning.
func Warnf(format string, a ...any) { Warn(fmt.Sprintf(format, a...)) }

// Errorf formats and prints a red error.
func Errorf(format string, a ...any) { Error(fmt.Sprintf(format, a...)) }

// Stepf formats and prints a bold step.
func Stepf(format string, a ...any) { Step(fmt.Sprintf(format, a...)) }
