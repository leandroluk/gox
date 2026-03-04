package di

import (
	"reflect"
	"sync"
	"sync/atomic"
)

// Provider holds the necessary information to create and manage an instance.
type Provider struct {
	Name            string        // Optional name for named instances (empty string for unnamed)
	FactoryFunction reflect.Value // The function used to create the instance.
	OutputType      reflect.Type  // The reflected type of the result.
	IsSingleton     bool          // Indicates if it should return the same instance every time.
	CachedInstance  reflect.Value // Stores the instance if it's a singleton.
	initOnce        sync.Once     // Ensures singleton is initialized only once.
	resolving       atomic.Bool   // Detects circular dependencies during resolution.

	// Lifecycle hooks (v1.4)
	OnStartHook  func(reflect.Value) error // Called when StartAll() is invoked
	OnStopHook   func(reflect.Value) error // Called when StopAll() is invoked
	hasLifecycle bool                      // Indicates if this provider has lifecycle hooks
	started      atomic.Bool               // Tracks if OnStart has been called
}

// Reset clears all registered providers.
// Primarily used for unit tests to ensure isolation.
func Reset() {
	RegistryMutex.Lock()
	ProviderRegistry = make(map[reflect.Type]map[string]*Provider)
	RegistryMutex.Unlock()

	lifecycleMutex.Lock()
	lifecycleProviders = nil
	lifecycleMutex.Unlock()
}
