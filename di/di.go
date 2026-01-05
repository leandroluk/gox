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
