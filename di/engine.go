package di

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// resolutionContext tracks the dependency chain during resolution.
type resolutionContext struct {
	chain []reflect.Type
}

// formatChain creates a human-readable dependency chain string.
func (ctx *resolutionContext) formatChain() string {
	if len(ctx.chain) == 0 {
		return "(empty)"
	}

	parts := make([]string, len(ctx.chain))
	for i, t := range ctx.chain {
		parts[i] = t.String()
	}
	return strings.Join(parts, " -> ")
}

// resolveByType finds the registered provider and triggers the build process.
// Resolves the unnamed (default) provider.
func resolveByType(targetType reflect.Type) reflect.Value {
	return resolveByTypeWithContext(targetType, nil)
}

// resolveByTypeNamed finds a named provider and triggers the build process.
func resolveByTypeNamed(targetType reflect.Type, name string) reflect.Value {
	return resolveByTypeNamedWithContext(targetType, name, nil)
}

// resolveByTypeWithContext finds the registered provider with dependency tracking.
// Resolves the unnamed (default) provider.
func resolveByTypeWithContext(targetType reflect.Type, ctx *resolutionContext) reflect.Value {
	if ctx == nil {
		ctx = &resolutionContext{chain: []reflect.Type{}}
	}

	LogDebug("Resolving %v (unnamed)", targetType)

	// Add current type to chain
	ctx.chain = append(ctx.chain, targetType)
	defer func() {
		ctx.chain = ctx.chain[:len(ctx.chain)-1]
	}()

	RegistryMutex.RLock()
	namedProviders := ProviderRegistry[targetType]
	RegistryMutex.RUnlock()

	// Look for unnamed provider (empty string key)
	if namedProviders == nil || namedProviders[""] == nil {
		chainStr := ctx.formatChain()
		Fail(fmt.Sprintf("di: no provider registered for type %v\n  dependency chain: %s\n  hint: did you forget to register a provider for this type?", targetType, chainStr))
	}

	return buildInstanceWithContext(namedProviders[""], ctx)
}

// resolveByTypeNamedWithContext finds a named provider with dependency tracking.
func resolveByTypeNamedWithContext(targetType reflect.Type, name string, ctx *resolutionContext) reflect.Value {
	if ctx == nil {
		ctx = &resolutionContext{chain: []reflect.Type{}}
	}

	LogDebug("Resolving %v (named: %s)", targetType, name)

	// Add current type to chain
	ctx.chain = append(ctx.chain, targetType)
	defer func() {
		ctx.chain = ctx.chain[:len(ctx.chain)-1]
	}()

	RegistryMutex.RLock()
	namedProviders := ProviderRegistry[targetType]
	RegistryMutex.RUnlock()

	if namedProviders == nil || namedProviders[name] == nil {
		chainStr := ctx.formatChain()
		Fail(fmt.Sprintf("di: no provider registered for type %v with name %q\n  dependency chain: %s\n  hint: did you forget to register a named provider?", targetType, name, chainStr))
	}

	return buildInstanceWithContext(namedProviders[name], ctx)
}

// buildInstance manages the lifecycle of the instance (Transient vs Singleton).
func buildInstance(providerInstance *Provider) reflect.Value {
	return buildInstanceWithContext(providerInstance, nil)
}

// buildInstanceWithContext manages the lifecycle with dependency tracking.
func buildInstanceWithContext(providerInstance *Provider, ctx *resolutionContext) reflect.Value {
	if providerInstance.IsSingleton {
		// Must check resolving BEFORE initOnce.Do to avoid deadlock on circular deps.
		// initOnce.Do uses an internal mutex; a re-entrant call on the same goroutine
		// would block forever waiting for the first Do to finish.
		if providerInstance.resolving.Load() {
			if ctx == nil {
				ctx = &resolutionContext{chain: []reflect.Type{providerInstance.OutputType}}
			}
			chainStr := ctx.formatChain()
			Fail(fmt.Sprintf("di: circular dependency detected for type %v\n  dependency chain: %s -> %v (circular)\n  hint: refactor your dependencies to break the cycle",
				providerInstance.OutputType, chainStr, providerInstance.OutputType))
		}
		providerInstance.initOnce.Do(func() {
			providerInstance.CachedInstance = callFactoryWithDependencies(providerInstance, ctx)
		})
		return providerInstance.CachedInstance
	}

	return callFactoryWithDependencies(providerInstance, ctx)
}

// callFactoryWithDependencies recursively resolves all inputs of a factory function.
func callFactoryWithDependencies(providerInstance *Provider, ctx *resolutionContext) reflect.Value {
	// Detect circular dependency
	if providerInstance.resolving.Load() {
		if ctx == nil {
			ctx = &resolutionContext{chain: []reflect.Type{providerInstance.OutputType}}
		}
		chainStr := ctx.formatChain()
		Fail(fmt.Sprintf("di: circular dependency detected for type %v\n  dependency chain: %s -> %v (circular)\n  hint: refactor your dependencies to break the cycle",
			providerInstance.OutputType, chainStr, providerInstance.OutputType))
	}

	// Mark as resolving
	providerInstance.resolving.Store(true)
	defer providerInstance.resolving.Store(false)

	factoryType := providerInstance.FactoryFunction.Type()
	numberOfInputs := factoryType.NumIn()

	arguments := make([]reflect.Value, numberOfInputs)

	for index := 0; index < numberOfInputs; index++ {
		dependencyType := factoryType.In(index)
		// Dependencies are always resolved as unnamed (default behavior)
		arguments[index] = resolveByTypeWithContext(dependencyType, ctx)
	}

	outputValues := providerInstance.FactoryFunction.Call(arguments)
	return outputValues[0]
}

var (
	RegistryMutex    sync.RWMutex
	ProviderRegistry = map[reflect.Type]map[string]*Provider{}
)
