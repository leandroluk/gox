package di

import (
	"reflect"
)

// Register adds a transient provider (a new instance is created every time it's resolved).
func Register(factoryFN any) {
	registerProvider(factoryFN, false, nil, "")
}

// RegisterAs adds a transient provider bound to a specific interface or type T.
func RegisterAs[T any](factoryFN any) {
	registerProvider(factoryFN, false, reflect.TypeFor[T](), "")
}

// RegisterNamed adds a named transient provider.
func RegisterNamed[T any](name string, factoryFN any) {
	if name == "" {
		Fail("di: named registration requires a non-empty name")
	}
	registerProvider(factoryFN, false, reflect.TypeFor[T](), name)
}

// Singleton adds a provider that caches its instance after the first resolution.
func Singleton(factoryFN any) {
	registerProvider(factoryFN, true, nil, "")
}

// SingletonAs adds a singleton provider bound to a specific interface or type T.
func SingletonAs[T any](factoryFN any) {
	registerProvider(factoryFN, true, reflect.TypeFor[T](), "")
}

// SingletonNamed adds a named singleton provider.
func SingletonNamed[T any](name string, factoryFN any) {
	if name == "" {
		Fail("di: named registration requires a non-empty name")
	}
	registerProvider(factoryFN, true, reflect.TypeFor[T](), name)
}

// SingletonInstance adds an existing instance as a singleton provider for type T.
func SingletonInstance[T any](instance T) {
	registerProvider(func() T { return instance }, true, reflect.TypeFor[T](), "")
}

// SingletonInstanceNamed adds an existing instance as a named singleton provider.
func SingletonInstanceNamed[T any](name string, instance T) {
	if name == "" {
		Fail("di: named registration requires a non-empty name")
	}
	registerProvider(func() T { return instance }, true, reflect.TypeFor[T](), name)
}

// RegisterWithLifecycle registers a provider with lifecycle hooks.
// The factory must return a type, and hooks are called during StartAll/StopAll.
//
// Example:
//
//	di.RegisterWithLifecycle(NewDatabase, di.LifecycleHooks{
//	    OnStart: func(db any) error {
//	        return db.(*Database).Connect()
//	    },
//	    OnStop: func(db any) error {
//	        return db.(*Database).Close()
//	    },
//	})
func RegisterWithLifecycle(factoryFN any, hooks LifecycleHooks) {
	registerProviderWithLifecycle(factoryFN, false, nil, "", &hooks)
}

// RegisterAsWithLifecycle registers a typed provider with lifecycle hooks.
func RegisterAsWithLifecycle[T any](factoryFN any, hooks LifecycleHooks) {
	registerProviderWithLifecycle(factoryFN, false, reflect.TypeFor[T](), "", &hooks)
}

// RegisterNamedWithLifecycle registers a named provider with lifecycle hooks.
func RegisterNamedWithLifecycle[T any](name string, factoryFN any, hooks LifecycleHooks) {
	if name == "" {
		Fail("di: named registration requires a non-empty name")
	}
	registerProviderWithLifecycle(factoryFN, false, reflect.TypeFor[T](), name, &hooks)
}

// SingletonWithLifecycle registers a singleton provider with lifecycle hooks.
//
// Example:
//
//	di.SingletonWithLifecycle(NewDatabase, di.LifecycleHooks{
//	    OnStart: func(db any) error { return db.(*Database).Connect() },
//	    OnStop:  func(db any) error { return db.(*Database).Close() },
//	})
func SingletonWithLifecycle(factoryFN any, hooks LifecycleHooks) {
	registerProviderWithLifecycle(factoryFN, true, nil, "", &hooks)
}

// SingletonAsWithLifecycle registers a typed singleton with lifecycle hooks.
func SingletonAsWithLifecycle[T any](factoryFN any, hooks LifecycleHooks) {
	registerProviderWithLifecycle(factoryFN, true, reflect.TypeFor[T](), "", &hooks)
}

// SingletonNamedWithLifecycle registers a named singleton with lifecycle hooks.
func SingletonNamedWithLifecycle[T any](name string, factoryFN any, hooks LifecycleHooks) {
	if name == "" {
		Fail("di: named registration requires a non-empty name")
	}
	registerProviderWithLifecycle(factoryFN, true, reflect.TypeFor[T](), name, &hooks)
}

// Resolve retrieves the unnamed (default) instance for type T. Panics if no provider is found.
func Resolve[T any]() T {
	targetType := reflect.TypeFor[T]()
	value := resolveByType(targetType)
	return value.Interface().(T)
}

// ResolveNamed retrieves a named instance for type T.
func ResolveNamed[T any](name string) T {
	if name == "" {
		Fail("di: ResolveNamed requires a non-empty name")
	}
	targetType := reflect.TypeFor[T]()
	value := resolveByTypeNamed(targetType, name)
	return value.Interface().(T)
}

// TryResolve attempts to resolve the unnamed type T without panicking.
func TryResolve[T any]() (T, bool) {
	var zero T
	targetType := reflect.TypeFor[T]()

	RegistryMutex.RLock()
	namedProviders := ProviderRegistry[targetType]
	RegistryMutex.RUnlock()

	if namedProviders == nil || namedProviders[""] == nil {
		LogDebug("TryResolve: no unnamed provider found for %v", targetType)
		return zero, false
	}

	LogDebug("TryResolve: found unnamed provider for %v", targetType)
	value := buildInstance(namedProviders[""])
	return value.Interface().(T), true
}

// TryResolveNamed attempts to resolve a named type T without panicking.
func TryResolveNamed[T any](name string) (T, bool) {
	var zero T
	if name == "" {
		LogDebug("TryResolveNamed: empty name provided")
		return zero, false
	}

	targetType := reflect.TypeFor[T]()

	RegistryMutex.RLock()
	namedProviders := ProviderRegistry[targetType]
	RegistryMutex.RUnlock()

	if namedProviders == nil || namedProviders[name] == nil {
		LogDebug("TryResolveNamed: no provider found for %v with name %q", targetType, name)
		return zero, false
	}

	LogDebug("TryResolveNamed: found provider for %v with name %q", targetType, name)
	value := buildInstance(namedProviders[name])
	return value.Interface().(T), true
}

// MustResolve resolves the unnamed type T or panics with a custom error message.
func MustResolve[T any](customMessage string) T {
	targetType := reflect.TypeFor[T]()

	RegistryMutex.RLock()
	namedProviders := ProviderRegistry[targetType]
	RegistryMutex.RUnlock()

	if namedProviders == nil || namedProviders[""] == nil {
		Fail(customMessage)
	}

	value := buildInstance(namedProviders[""])
	return value.Interface().(T)
}

// MustResolveNamed resolves a named type T or panics with a custom error message.
func MustResolveNamed[T any](name string, customMessage string) T {
	if name == "" {
		Fail("di: MustResolveNamed requires a non-empty name")
	}

	targetType := reflect.TypeFor[T]()

	RegistryMutex.RLock()
	namedProviders := ProviderRegistry[targetType]
	RegistryMutex.RUnlock()

	if namedProviders == nil || namedProviders[name] == nil {
		Fail(customMessage)
	}

	value := buildInstance(namedProviders[name])
	return value.Interface().(T)
}

// ResolveAll retrieves all registered providers for type T as a slice.
func ResolveAll[T any]() []T {
	targetType := reflect.TypeFor[T]()

	RegistryMutex.RLock()
	namedProviders := ProviderRegistry[targetType]
	RegistryMutex.RUnlock()

	if len(namedProviders) == 0 {
		return nil
	}

	results := make([]T, 0, len(namedProviders))
	for _, providerInstance := range namedProviders {
		value := buildInstance(providerInstance)
		results = append(results, value.Interface().(T))
	}

	return results
}

// ResolveAllNamed retrieves all named providers for type T as a map[name]instance.
func ResolveAllNamed[T any]() map[string]T {
	targetType := reflect.TypeFor[T]()

	RegistryMutex.RLock()
	namedProviders := ProviderRegistry[targetType]
	RegistryMutex.RUnlock()

	if len(namedProviders) == 0 {
		return nil
	}

	results := make(map[string]T)
	for name, providerInstance := range namedProviders {
		if name == "" {
			continue
		}
		value := buildInstance(providerInstance)
		results[name] = value.Interface().(T)
	}

	if len(results) == 0 {
		return nil
	}

	return results
}

// TryResolveAll attempts to resolve all instances of type T without panicking.
func TryResolveAll[T any]() ([]T, bool) {
	targetType := reflect.TypeFor[T]()

	RegistryMutex.RLock()
	namedProviders := ProviderRegistry[targetType]
	RegistryMutex.RUnlock()

	if len(namedProviders) == 0 {
		LogDebug("TryResolveAll: no providers found for %v", targetType)
		return nil, false
	}

	LogDebug("TryResolveAll: found %d provider(s) for %v", len(namedProviders), targetType)
	results := make([]T, 0, len(namedProviders))
	for _, providerInstance := range namedProviders {
		value := buildInstance(providerInstance)
		results = append(results, value.Interface().(T))
	}

	return results, true
}

// TryResolveAllNamed attempts to resolve all named instances of type T without panicking.
func TryResolveAllNamed[T any]() (map[string]T, bool) {
	targetType := reflect.TypeFor[T]()

	RegistryMutex.RLock()
	namedProviders := ProviderRegistry[targetType]
	RegistryMutex.RUnlock()

	if len(namedProviders) == 0 {
		LogDebug("TryResolveAllNamed: no providers found for %v", targetType)
		return nil, false
	}

	results := make(map[string]T)
	for name, providerInstance := range namedProviders {
		if name == "" {
			continue
		}
		value := buildInstance(providerInstance)
		results[name] = value.Interface().(T)
	}

	if len(results) == 0 {
		LogDebug("TryResolveAllNamed: no named providers found for %v", targetType)
		return nil, false
	}

	LogDebug("TryResolveAllNamed: found %d named provider(s) for %v", len(results), targetType)
	return results, true
}
