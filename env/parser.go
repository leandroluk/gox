package env

import (
	"bufio"
	"os"
	"strings"
)

// loadEnvFile reads a file line by line and populates the store.
func loadEnvFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key, value, ok := parseEnvLine(scanner.Text())
		if !ok {
			continue
		}

		expandedValue := expandValue(value)

		envMutex.Lock()
		envStore[key] = expandedValue
		envMutex.Unlock()

		_ = os.Setenv(key, expandedValue)
	}
	return scanner.Err()
}

// parseEnvLine extracts key and value, handling spaces, quotes and comments.
func parseEnvLine(rawLine string) (string, string, bool) {
	line := strings.TrimSpace(rawLine)
	if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
		return "", "", false
	}

	if strings.HasPrefix(line, "export ") {
		line = strings.TrimSpace(line[len("export "):])
	}

	equalIndex := strings.Index(line, "=")
	if equalIndex < 0 {
		return "", "", false
	}

	key := strings.TrimSpace(line[:equalIndex])
	value := strings.TrimSpace(line[equalIndex+1:])
	value = stripInlineComments(value)

	// Remove surrounding quotes
	if len(value) >= 2 {
		first, last := value[0], value[len(value)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			value = value[1 : len(value)-1]
		}
	}

	if key == "" {
		return "", "", false
	}
	return key, value, true
}

func stripInlineComments(value string) string {
	inQuotes := false
	var quoteChar rune

	for i, char := range value {
		if char == '"' || char == '\'' {
			if !inQuotes {
				inQuotes, quoteChar = true, char
			} else if quoteChar == char {
				inQuotes = false
			}
		}
		if inQuotes {
			continue
		}
		if char == '#' || (char == '/' && i+1 < len(value) && value[i+1] == '/') {
			return strings.TrimSpace(value[:i])
		}
	}
	return strings.TrimSpace(value)
}
