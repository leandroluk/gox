package di

import "fmt"

var DebugMode = false

// Debug enables debug mode, which prints registration and resolution events,
// and prints errors before panicking.
func Debug() {
	DebugMode = true
}

// LogDebug prints a formatted message if debug mode is enabled.
func LogDebug(format string, args ...any) {
	if DebugMode {
		fmt.Printf("[DI] "+format+"\n", args...)
	}
}

// Fail prints the error message if debug mode is enabled, then panics.
func Fail(msg string) {
	if DebugMode {
		fmt.Printf("[DI] ERROR: %s\n", msg)
	}
	panic(msg)
}
