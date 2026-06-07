package di

import "reflect"

func Register[T any](configurator func(*Options[T])) {
	opts := &Options[T]{}
	if configurator != nil {
		configurator(opts)
	}
	registerProvider(opts, false, reflect.TypeFor[T](), "")
}

func RegisterAs[T any](configurator func(*Options[T])) {
	opts := &Options[T]{}
	if configurator != nil {
		configurator(opts)
	}
	registerProvider(opts, false, reflect.TypeFor[T](), "")
}

func RegisterNamed[T any](name string, configurator func(*Options[T])) {
	if name == "" {
		Fail("di: named registration requires a non-empty name")
	}
	opts := &Options[T]{}
	if configurator != nil {
		configurator(opts)
	}
	registerProvider(opts, false, reflect.TypeFor[T](), name)
}

func Singleton[T any](configurator func(*Options[T])) {
	opts := &Options[T]{}
	if configurator != nil {
		configurator(opts)
	}
	registerProvider(opts, true, reflect.TypeFor[T](), "")
}

func SingletonAs[T any](configurator func(*Options[T])) {
	opts := &Options[T]{}
	if configurator != nil {
		configurator(opts)
	}
	registerProvider(opts, true, reflect.TypeFor[T](), "")
}

func SingletonNamed[T any](name string, configurator func(*Options[T])) {
	if name == "" {
		Fail("di: named registration requires a non-empty name")
	}
	opts := &Options[T]{}
	if configurator != nil {
		configurator(opts)
	}
	registerProvider(opts, true, reflect.TypeFor[T](), name)
}

func SingletonInstance[T any](instance T, configurator func(*Options[T])) {
	opts := &Options[T]{}
	opts.Constructor = func() (T, error) { return instance, nil }
	if configurator != nil {
		configurator(opts)
	}
	registerProvider(opts, true, reflect.TypeFor[T](), "")
}

func SingletonInstanceNamed[T any](name string, instance T, configurator func(*Options[T])) {
	if name == "" {
		Fail("di: named registration requires a non-empty name")
	}
	opts := &Options[T]{}
	opts.Constructor = func() (T, error) { return instance, nil }
	if configurator != nil {
		configurator(opts)
	}
	registerProvider(opts, true, reflect.TypeFor[T](), name)
}

func SingletonFrom[T any](constructor func() (T, error)) {
	opts := &Options[T]{Constructor: constructor}
	registerProvider(opts, true, reflect.TypeFor[T](), "")
}

func RegisterFrom[T any](constructor func() (T, error)) {
	opts := &Options[T]{Constructor: constructor}
	registerProvider(opts, false, reflect.TypeFor[T](), "")
}

func Resolve[T any]() T {
	targetType := reflect.TypeFor[T]()
	value := resolveByType(targetType)
	return value.Interface().(T)
}

func ResolveNamed[T any](name string) T {
	if name == "" {
		Fail("di: ResolveNamed requires a non-empty name")
	}
	targetType := reflect.TypeFor[T]()
	value := resolveByTypeNamed(targetType, name)
	return value.Interface().(T)
}

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
