package di

import (
	"reflect"
	"sync"
)

var (
	registryMutex    sync.RWMutex
	providerRegistry = map[reflect.Type][]*Provider{}
)

// Provider holds the necessary information to create and manage an instance.
type Provider struct {
	FactoryFunction reflect.Value // The function used to create the instance.
	OutputType      reflect.Type  // The reflected type of the result.
	IsSingleton     bool          // Indicates if it should return the same instance every time.
	CachedInstance  reflect.Value // Stores the instance if it's a singleton.
}

// Reset clears all registered providers.
// Primarily used for unit tests to ensure isolation.
func Reset() {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	providerRegistry = make(map[reflect.Type][]*Provider)
}
