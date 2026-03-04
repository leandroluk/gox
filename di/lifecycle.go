package di

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

// LifecycleHooks defines the lifecycle callbacks for a provider.
type LifecycleHooks struct {
	OnStart func(instance any) error
	OnStop  func(instance any) error
}

// addLifecycleProvider adds a provider to the lifecycle management list.
func addLifecycleProvider(provider *Provider) {
	if !provider.hasLifecycle {
		return
	}

	lifecycleMutex.Lock()
	defer lifecycleMutex.Unlock()
	lifecycleProviders = append(lifecycleProviders, provider)

	LogDebug("Added lifecycle provider: %v (has OnStart: %v, has OnStop: %v)",
		provider.OutputType, provider.OnStartHook != nil, provider.OnStopHook != nil)
}

// StartAll initializes all providers with OnStart hooks.
// Providers are started in registration order.
// If any OnStart fails, all previously started providers are stopped (rollback).
func StartAll() error {
	return StartAllWithContext(context.Background())
}

// StartAllWithContext initializes all providers with context support.
func StartAllWithContext(ctx context.Context) error {
	lifecycleMutex.RLock()
	providers := make([]*Provider, len(lifecycleProviders))
	copy(providers, lifecycleProviders)
	lifecycleMutex.RUnlock()

	LogDebug("Starting %d lifecycle provider(s)", len(providers))

	var startedProviders []*Provider

	for i, provider := range providers {
		// Check context cancellation
		select {
		case <-ctx.Done():
			LogDebug("StartAll cancelled by context")
			// Rollback what we started
			rollbackStart(startedProviders)
			return ctx.Err()
		default:
		}

		if provider.OnStartHook == nil {
			continue
		}

		// Skip if already started
		if provider.started.Load() {
			LogDebug("Provider %v already started, skipping", provider.OutputType)
			continue
		}

		LogDebug("Starting provider %d/%d: %v", i+1, len(providers), provider.OutputType)

		// Get or build instance
		var instance reflect.Value
		if provider.IsSingleton && provider.CachedInstance.IsValid() {
			instance = provider.CachedInstance
		} else {
			instance = buildInstance(provider)
		}

		// Call OnStart hook with context cancellation support
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

// rollbackStart stops all providers that were started, in reverse order.
func rollbackStart(providers []*Provider) {
	LogDebug("Rolling back %d started provider(s)", len(providers))

	// Stop in reverse order
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
			continue // Can't stop non-singleton that wasn't cached
		}

		if err := provider.OnStopHook(instance); err != nil {
			LogDebug("Error during rollback of %v: %v", provider.OutputType, err)
			// Continue rollback even if one fails
		}

		provider.started.Store(false)
	}
}

// StopAll stops all providers with OnStop hooks.
// Providers are stopped in reverse registration order (LIFO).
// All OnStop hooks are called even if some fail.
func StopAll() error {
	return StopAllWithContext(context.Background())
}

// StopAllWithContext stops all providers with context support.
func StopAllWithContext(ctx context.Context) error {
	lifecycleMutex.RLock()
	providers := make([]*Provider, len(lifecycleProviders))
	copy(providers, lifecycleProviders)
	lifecycleMutex.RUnlock()

	LogDebug("Stopping %d lifecycle provider(s)", len(providers))

	var errs []error

	// Stop in reverse order (LIFO)
	for i := len(providers) - 1; i >= 0; i-- {
		provider := providers[i]

		// Check context cancellation
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

		// Skip if has OnStart but not started
		if provider.OnStartHook != nil && !provider.started.Load() {
			LogDebug("Provider %v not started, skipping stop", provider.OutputType)
			continue
		}

		LogDebug("Stopping provider %d/%d: %v", len(providers)-i, len(providers), provider.OutputType)

		// Get instance
		var instance reflect.Value
		if provider.IsSingleton && provider.CachedInstance.IsValid() {
			instance = provider.CachedInstance
		} else {
			LogDebug("Cannot stop %v: no cached instance", provider.OutputType)
			continue
		}

		// Call OnStop hook
		if err := provider.OnStopHook(instance); err != nil {
			LogDebug("Error stopping %v: %v", provider.OutputType, err)
			errs = append(errs, fmt.Errorf("%v: %w", provider.OutputType, err))
			// Continue stopping others even if one fails
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

// StartAllWithTimeout starts all providers with a timeout.
func StartAllWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return StartAllWithContext(ctx)
}

// StopAllWithTimeout stops all providers with a timeout.
func StopAllWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return StopAllWithContext(ctx)
}

// GetLifecycleProviders returns the list of providers with lifecycle hooks (for testing).
func GetLifecycleProviders() []*Provider {
	lifecycleMutex.RLock()
	defer lifecycleMutex.RUnlock()

	providers := make([]*Provider, len(lifecycleProviders))
	copy(providers, lifecycleProviders)
	return providers
}
