package di_v1

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"
)

var (
	lifecycleProviders []*Provider
	lifecycleMutex     sync.RWMutex
)

func addLifecycleProvider(provider *Provider) {
	if !provider.hasLifecycle {
		return
	}

	lifecycleMutex.Lock()
	defer lifecycleMutex.Unlock()
	lifecycleProviders = append(lifecycleProviders, provider)

	LogDebug("Added lifecycle provider: %v (has OnApplicationBootstrap: %v, has OnApplicationShutdown: %v)",
		provider.OutputType, provider.OnStartHook != nil, provider.OnStopHook != nil)
}

func StartAll() error {
	return StartAllWithContext(context.Background())
}

func StartAllWithContext(ctx context.Context) error {
	lifecycleMutex.RLock()
	providers := make([]*Provider, len(lifecycleProviders))
	copy(providers, lifecycleProviders)
	lifecycleMutex.RUnlock()

	LogDebug("Starting %d lifecycle provider(s)", len(providers))

	var startedProviders []*Provider

	for i, provider := range providers {
		select {
		case <-ctx.Done():
			LogDebug("StartAll cancelled by context")
			rollbackStart(startedProviders)
			return ctx.Err()
		default:
		}

		if provider.OnStartHook == nil {
			continue
		}

		if provider.started.Load() {
			LogDebug("Provider %v already started, skipping", provider.OutputType)
			continue
		}

		LogDebug("Starting provider %d/%d: %v", i+1, len(providers), provider.OutputType)

		var instance reflect.Value
		if provider.IsSingleton && provider.CachedInstance.IsValid() {
			instance = provider.CachedInstance
		} else {
			instance = buildInstance(provider)
		}

		type hookResult struct{ err error }
		done := make(chan hookResult, 1)
		go func() {
			done <- hookResult{err: provider.OnStartHook(instance)}
		}()

		select {
		case <-ctx.Done():
			LogDebug("StartAll cancelled by context during hook execution")
			rollbackStart(startedProviders)
			return ctx.Err()
		case res := <-done:
			if res.err != nil {
				LogDebug("Failed to start %v: %v", provider.OutputType, res.err)
				rollbackStart(startedProviders)
				return fmt.Errorf("failed to start %v: %w", provider.OutputType, res.err)
			}
		}

		provider.started.Store(true)
		startedProviders = append(startedProviders, provider)
		LogDebug("Successfully started: %v", provider.OutputType)
	}

	LogDebug("All providers started successfully")
	return nil
}

func rollbackStart(providers []*Provider) {
	LogDebug("Rolling back %d started provider(s)", len(providers))

	for i := len(providers) - 1; i >= 0; i-- {
		provider := providers[i]
		if provider.OnStopHook == nil {
			continue
		}

		LogDebug("Rolling back: %v", provider.OutputType)

		var instance reflect.Value
		if provider.IsSingleton && provider.CachedInstance.IsValid() {
			instance = provider.CachedInstance
		} else {
			continue
		}

		if err := provider.OnStopHook(instance); err != nil {
			LogDebug("Error during rollback of %v: %v", provider.OutputType, err)
		}

		provider.started.Store(false)
	}
}

func StopAll() error {
	return StopAllWithContext(context.Background())
}

func StopAllWithContext(ctx context.Context) error {
	lifecycleMutex.RLock()
	providers := make([]*Provider, len(lifecycleProviders))
	copy(providers, lifecycleProviders)
	lifecycleMutex.RUnlock()

	LogDebug("Stopping %d lifecycle provider(s)", len(providers))

	var errs []error

	for i := len(providers) - 1; i >= 0; i-- {
		provider := providers[i]

		select {
		case <-ctx.Done():
			LogDebug("StopAll cancelled by context")
			if len(errs) > 0 {
				return fmt.Errorf("stop cancelled with %d error(s): first error: %w", len(errs), errs[0])
			}
			return ctx.Err()
		default:
		}

		if provider.OnStopHook == nil {
			continue
		}

		if provider.OnStartHook != nil && !provider.started.Load() {
			LogDebug("Provider %v not started, skipping stop", provider.OutputType)
			continue
		}

		LogDebug("Stopping provider %d/%d: %v", len(providers)-i, len(providers), provider.OutputType)

		var instance reflect.Value
		if provider.IsSingleton && provider.CachedInstance.IsValid() {
			instance = provider.CachedInstance
		} else {
			LogDebug("Cannot stop %v: no cached instance", provider.OutputType)
			continue
		}

		if err := provider.OnStopHook(instance); err != nil {
			LogDebug("Error stopping %v: %v", provider.OutputType, err)
			errs = append(errs, fmt.Errorf("%v: %w", provider.OutputType, err))
		} else {
			LogDebug("Successfully stopped: %v", provider.OutputType)
		}

		provider.started.Store(false)
	}

	if len(errs) > 0 {
		LogDebug("StopAll completed with %d error(s)", len(errs))
		return fmt.Errorf("stop completed with %d error(s): first error: %w", len(errs), errs[0])
	}

	LogDebug("All providers stopped successfully")
	return nil
}

func StartAllWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return StartAllWithContext(ctx)
}

func StopAllWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return StopAllWithContext(ctx)
}

func GetLifecycleProviders() []*Provider {
	lifecycleMutex.RLock()
	defer lifecycleMutex.RUnlock()

	providers := make([]*Provider, len(lifecycleProviders))
	copy(providers, lifecycleProviders)
	return providers
}
