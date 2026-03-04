package di_test

import (
	"strings"
	"testing"

	"github.com/leandroluk/gox/di"
)

// --- TryResolve Test Types (v1.2) ---

type TryLogger interface {
	Log(msg string)
}

type TryConsoleLogger struct {
	Prefix string
}

func (l *TryConsoleLogger) Log(msg string) {
	// Mock implementation
}

type TryFileLogger struct {
	FilePath string
}

func (l *TryFileLogger) Log(msg string) {
	// Mock implementation
}

type TryCache interface {
	Get(key string) string
	Set(key, value string)
}

type TryRedisCache struct {
	Host string
}

func (c *TryRedisCache) Get(key string) string { return "" }
func (c *TryRedisCache) Set(key, value string) {}

type TryConfig struct {
	AppName string
	Port    int
}

type TryPlugin interface {
	Name() string
	Execute()
}

type TryPluginA struct{}

func (p *TryPluginA) Name() string { return "A" }
func (p *TryPluginA) Execute()     {}

type TryPluginB struct{}

func (p *TryPluginB) Name() string { return "B" }
func (p *TryPluginB) Execute()     {}

// --- TryResolve Tests ---

func TestDI_TryResolve_Success(t *testing.T) {
	di.Reset()

	// Register a provider
	di.Register(func() *TryConfig {
		return &TryConfig{AppName: "TestApp", Port: 8080}
	})

	// Try to resolve - should succeed
	cfg, ok := di.TryResolve[*TryConfig]()

	if !ok {
		t.Fatal("TryResolve should succeed when provider is registered")
	}

	if cfg == nil {
		t.Fatal("TryResolve should return non-nil instance")
	}

	if cfg.AppName != "TestApp" {
		t.Errorf("Expected AppName 'TestApp', got '%s'", cfg.AppName)
	}

	if cfg.Port != 8080 {
		t.Errorf("Expected Port 8080, got %d", cfg.Port)
	}
}

func TestDI_TryResolve_NotFound(t *testing.T) {
	di.Reset()

	type UnregisteredType struct{}

	// Try to resolve unregistered type - should fail gracefully
	instance, ok := di.TryResolve[*UnregisteredType]()

	if ok {
		t.Error("TryResolve should return false for unregistered type")
	}

	if instance != nil {
		t.Error("TryResolve should return nil instance when not found")
	}

	// Should NOT panic - this is the whole point!
}

func TestDI_TryResolve_OptionalDependency(t *testing.T) {
	di.Reset()

	// Simulate optional logger
	// Logger is NOT registered

	// Use TryResolve to handle gracefully
	if logger, ok := di.TryResolve[TryLogger](); ok {
		logger.Log("test")
		t.Error("Should not have found logger")
	} else {
		// Expected path - no logger, use fallback
		// In real code: fmt.Println("test")
	}

	// Now register logger and try again
	di.RegisterAs[TryLogger](func() *TryConsoleLogger {
		return &TryConsoleLogger{Prefix: "[TEST]"}
	})

	if logger, ok := di.TryResolve[TryLogger](); ok {
		logger.Log("test") // Should work now
	} else {
		t.Error("Should have found logger after registration")
	}
}

func TestDI_TryResolve_WithSingleton(t *testing.T) {
	di.Reset()

	// Register as singleton
	di.Singleton(func() *TryConfig {
		return &TryConfig{AppName: "Singleton", Port: 3000}
	})

	// Try resolve multiple times - should get same instance
	cfg1, ok1 := di.TryResolve[*TryConfig]()
	cfg2, ok2 := di.TryResolve[*TryConfig]()

	if !ok1 || !ok2 {
		t.Fatal("TryResolve should succeed for singleton")
	}

	// Should be same instance
	if cfg1 != cfg2 {
		t.Error("TryResolve should return same singleton instance")
	}
}

func TestDI_TryResolve_WithDependencies(t *testing.T) {
	di.Reset()

	type TryDatabase struct {
		Config *TryConfig
	}

	// Register config and database
	di.Register(func() *TryConfig {
		return &TryConfig{AppName: "DB App", Port: 5432}
	})

	di.Register(func(cfg *TryConfig) *TryDatabase {
		return &TryDatabase{Config: cfg}
	})

	// Try resolve database - should resolve with dependencies
	db, ok := di.TryResolve[*TryDatabase]()

	if !ok {
		t.Fatal("TryResolve should succeed when dependencies are satisfied")
	}

	if db.Config == nil {
		t.Error("Dependencies should be injected")
	}

	if db.Config.AppName != "DB App" {
		t.Error("Dependency values should be correct")
	}
}

// --- MustResolve Tests ---

func TestDI_MustResolve_Success(t *testing.T) {
	di.Reset()

	di.Register(func() *TryConfig {
		return &TryConfig{AppName: "MustApp", Port: 9000}
	})

	// Should not panic
	cfg := di.MustResolve[*TryConfig]("Config is required!")

	if cfg == nil {
		t.Fatal("MustResolve should return instance")
	}

	if cfg.AppName != "MustApp" {
		t.Error("MustResolve should return correct instance")
	}
}

func TestDI_MustResolve_CustomMessage(t *testing.T) {
	di.Reset()

	customMsg := "Database configuration is required for application startup"

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("MustResolve should panic when provider not found")
		}

		errMsg := r.(string)
		if errMsg != customMsg {
			t.Errorf("Expected custom message '%s', got '%s'", customMsg, errMsg)
		}
	}()

	type TryDatabase struct{}
	di.MustResolve[*TryDatabase](customMsg)
}

func TestDI_MustResolve_VsResolve(t *testing.T) {
	di.Reset()

	type TryService struct{}

	// Test that MustResolve with custom message is different from Resolve
	t.Run("Resolve default message", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("Resolve should panic")
			}

			errMsg := r.(string)
			// Default message includes type info
			if !strings.Contains(errMsg, "TryService") {
				t.Error("Default message should mention type")
			}
		}()

		di.Resolve[*TryService]()
	})

	t.Run("MustResolve custom message", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("MustResolve should panic")
			}

			errMsg := r.(string)
			if errMsg != "Service layer is not configured" {
				t.Error("Should use custom message")
			}
		}()

		di.MustResolve[*TryService]("Service layer is not configured")
	})
}

// --- TryResolveAll Tests ---

func TestDI_TryResolveAll_Success(t *testing.T) {
	di.Reset()

	// Register multiple loggers using named instances (v1.3)
	di.RegisterNamed[TryLogger]("console", func() *TryConsoleLogger {
		return &TryConsoleLogger{Prefix: "[CONSOLE]"}
	})

	di.RegisterNamed[TryLogger]("file", func() *TryFileLogger {
		return &TryFileLogger{FilePath: "/var/log/app.log"}
	})

	// Try resolve all - should succeed
	loggers, ok := di.TryResolveAll[TryLogger]()

	if !ok {
		t.Fatal("TryResolveAll should succeed when providers exist")
	}

	if len(loggers) != 2 {
		t.Errorf("Expected 2 loggers, got %d", len(loggers))
	}
}

func TestDI_TryResolveAll_NotFound(t *testing.T) {
	di.Reset()

	type TryHandler interface {
		Handle()
	}

	// Try resolve all with no registrations
	handlers, ok := di.TryResolveAll[TryHandler]()

	if ok {
		t.Error("TryResolveAll should return false when no providers found")
	}

	if handlers != nil {
		t.Error("TryResolveAll should return nil when not found")
	}

	// Should NOT panic
}

func TestDI_TryResolveAll_Single(t *testing.T) {
	di.Reset()

	// Register single cache
	di.RegisterAs[TryCache](func() *TryRedisCache {
		return &TryRedisCache{Host: "localhost"}
	})

	// Try resolve all - should work with single provider
	caches, ok := di.TryResolveAll[TryCache]()

	if !ok {
		t.Fatal("TryResolveAll should succeed with single provider")
	}

	if len(caches) != 1 {
		t.Errorf("Expected 1 cache, got %d", len(caches))
	}
}

// --- Real-World Scenarios ---

func TestDI_RealWorld_FeatureFlags(t *testing.T) {
	di.Reset()

	type TryFeatureFlags interface {
		IsEnabled(feature string) bool
	}

	// Feature flags are optional - app works without them
	if flags, ok := di.TryResolve[TryFeatureFlags](); ok {
		if flags.IsEnabled("new-ui") {
			// Use new UI
		}
	} else {
		// No feature flags - use defaults
		// This is fine!
	}

	// Test passes - no panic even though TryFeatureFlags not registered
}

func TestDI_RealWorld_Plugins(t *testing.T) {
	di.Reset()

	// Register plugins using named instances (v1.3)
	di.RegisterNamed[TryPlugin]("pluginA", func() *TryPluginA { return &TryPluginA{} })
	di.RegisterNamed[TryPlugin]("pluginB", func() *TryPluginB { return &TryPluginB{} })

	// Load all available plugins
	if plugins, ok := di.TryResolveAll[TryPlugin](); ok {
		for _, plugin := range plugins {
			plugin.Execute()
		}
	}

	// If no plugins registered, app still works
	// Just with reduced functionality
}

func TestDI_RealWorld_GracefulDegradation(t *testing.T) {
	di.Reset()

	type TryMetricsCollector interface {
		Track(event string)
	}

	// Metrics are optional - app works without them
	var metrics TryMetricsCollector
	if m, ok := di.TryResolve[TryMetricsCollector](); ok {
		metrics = m
	}

	// Use metrics if available
	if metrics != nil {
		metrics.Track("user.login")
	}
	// Otherwise, silently continue

	// This pattern enables graceful degradation
}

// --- Performance Tests ---

func TestDI_TryResolve_Performance(t *testing.T) {
	di.Reset()

	di.Register(func() *TryConfig {
		return &TryConfig{AppName: "PerfTest"}
	})

	// TryResolve should be as fast as Resolve
	for i := 0; i < 1000; i++ {
		if _, ok := di.TryResolve[*TryConfig](); !ok {
			t.Fatal("Should resolve successfully")
		}
	}
}

// --- Edge Cases ---

func TestDI_TryResolve_ZeroValue(t *testing.T) {
	di.Reset()

	// For pointer types, zero value is nil
	ptr, ok := di.TryResolve[*TryConfig]()
	if ok || ptr != nil {
		t.Error("Should return (nil, false) for unregistered pointer type")
	}

	// For interface types, zero value is nil
	iface, ok := di.TryResolve[TryLogger]()
	if ok || iface != nil {
		t.Error("Should return (nil, false) for unregistered interface")
	}

	// For struct types, zero value is empty struct
	type SimpleStruct struct {
		Value int
	}
	s, ok := di.TryResolve[SimpleStruct]()
	if ok {
		t.Error("Should return false for unregistered struct")
	}
	if s.Value != 0 {
		t.Error("Should return zero value struct")
	}
}

func TestDI_TryResolve_AfterReset(t *testing.T) {
	di.Reset()

	// Register and resolve
	di.Register(func() *TryConfig { return &TryConfig{} })
	if _, ok := di.TryResolve[*TryConfig](); !ok {
		t.Fatal("Should resolve before reset")
	}

	// Reset registry
	di.Reset()

	// Should not resolve anymore
	if _, ok := di.TryResolve[*TryConfig](); ok {
		t.Error("Should not resolve after reset")
	}
}
