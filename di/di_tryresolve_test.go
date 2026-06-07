package di_test

import (
	"strings"
	"testing"

	"github.com/leandroluk/gox/di"
)

type TryLogger interface {
	Log(msg string)
}

type TryConsoleLogger struct {
	Prefix string
}

func (l *TryConsoleLogger) Log(msg string) {}

type TryFileLogger struct {
	FilePath string
}

func (l *TryFileLogger) Log(msg string) {}

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

func TestDI_TryResolve_Success(t *testing.T) {
	di.Reset()

	di.RegisterFrom[*TryConfig](func() (*TryConfig, error) {
		return &TryConfig{AppName: "TestApp", Port: 8080}, nil
	})

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

	instance, ok := di.TryResolve[*UnregisteredType]()

	if ok {
		t.Error("TryResolve should return false for unregistered type")
	}

	if instance != nil {
		t.Error("TryResolve should return nil instance when not found")
	}
}

func TestDI_TryResolve_OptionalDependency(t *testing.T) {
	di.Reset()

	if logger, ok := di.TryResolve[TryLogger](); ok {
		logger.Log("test")
		t.Error("Should not have found logger")
	}

	di.RegisterAs[TryLogger](func(o *di.Options[TryLogger]) {
		o.Constructor = func() (TryLogger, error) {
			return &TryConsoleLogger{Prefix: "[TEST]"}, nil
		}
	})

	if logger, ok := di.TryResolve[TryLogger](); ok {
		logger.Log("test")
	} else {
		t.Error("Should have found logger after registration")
	}
}

func TestDI_TryResolve_WithSingleton(t *testing.T) {
	di.Reset()

	di.SingletonFrom[*TryConfig](func() (*TryConfig, error) {
		return &TryConfig{AppName: "Singleton", Port: 3000}, nil
	})

	cfg1, ok1 := di.TryResolve[*TryConfig]()
	cfg2, ok2 := di.TryResolve[*TryConfig]()

	if !ok1 || !ok2 {
		t.Fatal("TryResolve should succeed for singleton")
	}

	if cfg1 != cfg2 {
		t.Error("TryResolve should return same singleton instance")
	}
}

func TestDI_TryResolve_WithDependencies(t *testing.T) {
	di.Reset()

	type TryDatabase struct {
		Config *TryConfig
	}

	di.RegisterFrom[*TryConfig](func() (*TryConfig, error) {
		return &TryConfig{AppName: "DB App", Port: 5432}, nil
	})

	di.RegisterFrom[*TryDatabase](func() (*TryDatabase, error) {
		cfg := di.Resolve[*TryConfig]()
		return &TryDatabase{Config: cfg}, nil
	})

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

func TestDI_MustResolve_Success(t *testing.T) {
	di.Reset()

	di.RegisterFrom[*TryConfig](func() (*TryConfig, error) {
		return &TryConfig{AppName: "MustApp", Port: 9000}, nil
	})

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

	t.Run("Resolve default message", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("Resolve should panic")
			}

			errMsg := r.(string)
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

func TestDI_TryResolveAll_Success(t *testing.T) {
	di.Reset()

	di.RegisterNamed[TryLogger]("console", func(o *di.Options[TryLogger]) {
		o.Constructor = func() (TryLogger, error) {
			return &TryConsoleLogger{Prefix: "[CONSOLE]"}, nil
		}
	})

	di.RegisterNamed[TryLogger]("file", func(o *di.Options[TryLogger]) {
		o.Constructor = func() (TryLogger, error) {
			return &TryFileLogger{FilePath: "/var/log/app.log"}, nil
		}
	})

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

	handlers, ok := di.TryResolveAll[TryHandler]()

	if ok {
		t.Error("TryResolveAll should return false when no providers found")
	}

	if handlers != nil {
		t.Error("TryResolveAll should return nil when not found")
	}
}

func TestDI_TryResolveAll_Single(t *testing.T) {
	di.Reset()

	di.RegisterAs[TryCache](func(o *di.Options[TryCache]) {
		o.Constructor = func() (TryCache, error) { return &TryRedisCache{Host: "localhost"}, nil }
	})

	caches, ok := di.TryResolveAll[TryCache]()

	if !ok {
		t.Fatal("TryResolveAll should succeed with single provider")
	}

	if len(caches) != 1 {
		t.Errorf("Expected 1 cache, got %d", len(caches))
	}
}

func TestDI_RealWorld_FeatureFlags(t *testing.T) {
	di.Reset()

	type TryFeatureFlags interface {
		IsEnabled(feature string) bool
	}

	if flags, ok := di.TryResolve[TryFeatureFlags](); ok {
		if flags.IsEnabled("new-ui") {
		}
	}
}

func TestDI_RealWorld_Plugins(t *testing.T) {
	di.Reset()

	di.RegisterNamed[TryPlugin]("pluginA", func(o *di.Options[TryPlugin]) {
		o.Constructor = func() (TryPlugin, error) { return &TryPluginA{}, nil }
	})
	di.RegisterNamed[TryPlugin]("pluginB", func(o *di.Options[TryPlugin]) {
		o.Constructor = func() (TryPlugin, error) { return &TryPluginB{}, nil }
	})

	if plugins, ok := di.TryResolveAll[TryPlugin](); ok {
		for _, plugin := range plugins {
			plugin.Execute()
		}
	}
}

func TestDI_RealWorld_GracefulDegradation(t *testing.T) {
	di.Reset()

	type TryMetricsCollector interface {
		Track(event string)
	}

	var metrics TryMetricsCollector
	if m, ok := di.TryResolve[TryMetricsCollector](); ok {
		metrics = m
	}

	if metrics != nil {
		metrics.Track("user.login")
	}
}

func TestDI_TryResolve_Performance(t *testing.T) {
	di.Reset()

	di.RegisterFrom[*TryConfig](func() (*TryConfig, error) {
		return &TryConfig{AppName: "PerfTest"}, nil
	})

	for i := 0; i < 1000; i++ {
		if _, ok := di.TryResolve[*TryConfig](); !ok {
			t.Fatal("Should resolve successfully")
		}
	}
}

func TestDI_TryResolve_ZeroValue(t *testing.T) {
	di.Reset()

	ptr, ok := di.TryResolve[*TryConfig]()
	if ok || ptr != nil {
		t.Error("Should return (nil, false) for unregistered pointer type")
	}

	iface, ok := di.TryResolve[TryLogger]()
	if ok || iface != nil {
		t.Error("Should return (nil, false) for unregistered interface")
	}

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

	di.RegisterFrom[*TryConfig](func() (*TryConfig, error) { return &TryConfig{}, nil })
	if _, ok := di.TryResolve[*TryConfig](); !ok {
		t.Fatal("Should resolve before reset")
	}

	di.Reset()

	if _, ok := di.TryResolve[*TryConfig](); ok {
		t.Error("Should not resolve after reset")
	}
}
