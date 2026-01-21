package di

import "reflect"

// Register adds a transient provider (a new instance is created every time it's resolved).
func Register(factoryFN any) {
	registerProvider(factoryFN, false, nil)
}

// RegisterAs adds a transient provider bound to a specific interface or type T.
func RegisterAs[T any](factoryFN any) {
	registerProvider(factoryFN, false, reflect.TypeFor[T]())
}

// Singleton adds a provider that caches its instance after the first resolution.
func Singleton(factoryFN any) {
	registerProvider(factoryFN, true, nil)
}

// SingletonAs adds a singleton provider bound to a specific interface or type T.
func SingletonAs[T any](factoryFN any) {
	registerProvider(factoryFN, true, reflect.TypeFor[T]())
}

// SingletonInstance adds an existing instance as a singleton provider for type T.
func SingletonInstance[T any](instance T) {
	registerProvider(func() T { return instance }, true, reflect.TypeFor[T]())
}

// Resolve retrieves the primary instance for type T. Panics if no provider is found.
func Resolve[T any]() T {
	targetType := reflect.TypeFor[T]()
	value := resolveByType(targetType)
	return value.Interface().(T)
}

// ResolveAll retrieves all registered providers for type T as a slice.
func ResolveAll[T any]() []T {
	targetType := reflect.TypeFor[T]()

	registryMutex.RLock()
	providers := providerRegistry[targetType]
	registryMutex.RUnlock()

	if len(providers) == 0 {
		return nil
	}

	results := make([]T, 0, len(providers))
	for _, providerInstance := range providers {
		value := buildInstance(providerInstance)
		results = append(results, value.Interface().(T))
	}

	return results
}

// RegisterFallback sets a global fallback handler for dynamic type resolution.
func RegisterFallback(handler func(requestType reflect.Type) (instance any, ok bool)) {
	registerFallbackHandler(handler)
}

// SingletonGeneric registers a singleton and a fallback handler that attempts to use this instance
// for ANY interface request where the method names match.
// This allows a Generic Type (e.g. Client[string]) to satisfy a Specific Interface (e.g. Publisher[Account])
// effectively bypassing strict type checking via Duck Typing. Use with caution.
func SingletonGeneric[T any](instance T) {
	// Register the exact match
	SingletonInstance(instance)

	// Register the "Duck Typing" fallback
	instanceVal := reflect.ValueOf(instance)

	registerFallbackHandler(func(requestType reflect.Type) (any, bool) {
		// Only attempt resolution for Interfaces
		if requestType.Kind() != reflect.Interface {
			return nil, false
		}

		// Optimization: Check if instance implements the interface directly first
		// (Already checked by DI normal flow, but good for sanity)

		// Duck Typing Check:
		// Verify if `instance` has all methods required by `requestType` (by Name only).
		// We ignore signature compatibility because that is the point of this feature.

		for i := 0; i < requestType.NumMethod(); i++ {
			method := requestType.Method(i)
			if !instanceVal.MethodByName(method.Name).IsValid() {
				// Method missing
				return nil, false
			}
		}

		// If all methods are present by name, we assume compatibility/shim.
		return instance, true
	})
}
