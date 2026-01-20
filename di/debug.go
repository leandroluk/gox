package di

import "fmt"

var debugMode = false

// Debug enables debug mode, which prints registration and resolution events,
// and prints errors before panicking.
func Debug() {
	debugMode = true
}

// logDebug prints a formatted message if debug mode is enabled.
func logDebug(format string, args ...any) {
	if debugMode {
		fmt.Printf("[DI] "+format+"\n", args...)
	}
}

// fail prints the error message if debug mode is enabled, then panics.
func fail(msg string) {
	if debugMode {
		fmt.Printf("[DI] ERROR: %s\n", msg)
	}
	panic(msg)
}
