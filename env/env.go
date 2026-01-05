package env

import (
	"os"
	"strings"
	"sync"
)

var (
	envMutex sync.RWMutex
	envStore = map[string]string{}
)

// Load reads the first valid .env file from the provided paths.
func Load(filePaths ...string) {
	for _, path := range filePaths {
		if path == "" {
			continue
		}
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			continue
		}

		if err := loadEnvFile(path); err == nil {
			return
		}
	}
}

// Get retrieves an environment variable and converts it to T.
// Returns default value if not found or conversion fails.
func Get[T any](key string, optionalDefault ...T) T {
	var zero T
	val, found := lookupEnv(key)
	if !found || strings.TrimSpace(val) == "" {
		if len(optionalDefault) > 0 {
			return optionalDefault[0]
		}
		return zero
	}

	converted, err := convertStringToType[T](val)
	if err != nil {
		if len(optionalDefault) > 0 {
			return optionalDefault[0]
		}
		return zero
	}
	return converted
}

func lookupEnv(key string) (string, bool) {
	envMutex.RLock()
	defer envMutex.RUnlock()
	if v, ok := envStore[key]; ok {
		return v, true
	}
	return os.LookupEnv(key)
}

func expandValue(raw string) string {
	return os.Expand(raw, func(k string) string {
		v, _ := lookupEnv(k)
		return v
	})
}
