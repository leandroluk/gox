package di_v1

import (
	"fmt"
	"reflect"
	"strings"
)

type resolutionContext struct {
	chain []reflect.Type
}

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

func resolveByType(targetType reflect.Type) reflect.Value {
	return resolveByTypeWithContext(targetType, nil)
}

func resolveByTypeNamed(targetType reflect.Type, name string) reflect.Value {
	return resolveByTypeNamedWithContext(targetType, name, nil)
}

func resolveByTypeWithContext(targetType reflect.Type, ctx *resolutionContext) reflect.Value {
	if ctx == nil {
		ctx = &resolutionContext{chain: []reflect.Type{}}
	}

	LogDebug("Resolving %v (unnamed)", targetType)

	ctx.chain = append(ctx.chain, targetType)
	defer func() {
		ctx.chain = ctx.chain[:len(ctx.chain)-1]
	}()

	RegistryMutex.RLock()
	namedProviders := ProviderRegistry[targetType]
	RegistryMutex.RUnlock()

	if namedProviders == nil || namedProviders[""] == nil {
		chainStr := ctx.formatChain()
		Fail(fmt.Sprintf("di: no provider registered for type %v\n  dependency chain: %s\n  hint: did you forget to register a provider for this type?", targetType, chainStr))
	}

	return buildInstanceWithContext(namedProviders[""], ctx)
}

func resolveByTypeNamedWithContext(targetType reflect.Type, name string, ctx *resolutionContext) reflect.Value {
	if ctx == nil {
		ctx = &resolutionContext{chain: []reflect.Type{}}
	}

	LogDebug("Resolving %v (named: %s)", targetType, name)

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

func buildInstance(providerInstance *Provider) reflect.Value {
	return buildInstanceWithContext(providerInstance, nil)
}

func buildInstanceWithContext(providerInstance *Provider, ctx *resolutionContext) reflect.Value {
	if providerInstance.IsSingleton {
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

func callFactoryWithDependencies(providerInstance *Provider, ctx *resolutionContext) reflect.Value {
	if providerInstance.resolving.Load() {
		if ctx == nil {
			ctx = &resolutionContext{chain: []reflect.Type{providerInstance.OutputType}}
		}
		chainStr := ctx.formatChain()
		Fail(fmt.Sprintf("di: circular dependency detected for type %v\n  dependency chain: %s -> %v (circular)\n  hint: refactor your dependencies to break the cycle",
			providerInstance.OutputType, chainStr, providerInstance.OutputType))
	}

	providerInstance.resolving.Store(true)
	defer providerInstance.resolving.Store(false)

	factoryType := providerInstance.FactoryFunction.Type()
	numberOfInputs := factoryType.NumIn()

	arguments := make([]reflect.Value, numberOfInputs)

	for index := range numberOfInputs {
		dependencyType := factoryType.In(index)
		arguments[index] = resolveByTypeWithContext(dependencyType, ctx)
	}

	outputValues := providerInstance.FactoryFunction.Call(arguments)
	if len(outputValues) == 2 && !outputValues[1].IsNil() {
		err := outputValues[1].Interface().(error)
		Fail(fmt.Sprintf("di: factory function returned error for type %v: %v", providerInstance.OutputType, err))
	}

	return outputValues[0]
}
