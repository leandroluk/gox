package di_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/leandroluk/gox/di"
)

// --- Lifecycle Test Types ---

type LifecycleDatabase struct {
	Connected bool
	Closed    bool
	Name      string
}

func (d *LifecycleDatabase) Connect() error {
	if d.Connected {
		return errors.New("already connected")
	}
	d.Connected = true
	return nil
}

func (d *LifecycleDatabase) Close() error {
	if d.Closed {
		return errors.New("already closed")
	}
	d.Closed = true
	d.Connected = false
	return nil
}

type LifecycleCache struct {
	Started bool
	Stopped bool
	Name    string
}

func (c *LifecycleCache) Start() error {
	c.Started = true
	return nil
}

func (c *LifecycleCache) Stop() error {
	c.Stopped = true
	return nil
}

type LifecycleServer struct {
	Running bool
	Port    int
}

func (s *LifecycleServer) Start() error {
	s.Running = true
	return nil
}

func (s *LifecycleServer) Stop() error {
	s.Running = false
	return nil
}

// --- Basic Lifecycle Tests ---

func TestDI_Lifecycle_BasicStartStop(t *testing.T) {
	di.Reset()

	di.SingletonWithLifecycle(func() *LifecycleDatabase {
		return &LifecycleDatabase{Name: "test"}
	}, di.LifecycleHooks{
		OnStart: func(instance any) error {
			return instance.(*LifecycleDatabase).Connect()
		},
		OnStop: func(instance any) error {
			return instance.(*LifecycleDatabase).Close()
		},
	})

	// Start all
	if err := di.StartAll(); err != nil {
		t.Fatalf("StartAll failed: %v", err)
	}

	// Get instance to verify
	db := di.Resolve[*LifecycleDatabase]()

	// After start
	if !db.Connected {
		t.Error("Database should be connected after StartAll")
	}

	// Stop all
	if err := di.StopAll(); err != nil {
		t.Fatalf("StopAll failed: %v", err)
	}

	// After stop
	if !db.Closed {
		t.Error("Database should be closed after StopAll")
	}
}

func TestDI_Lifecycle_MultipleProviders(t *testing.T) {
	di.Reset()

	// Register in order: db, cache, server
	di.SingletonWithLifecycle(func() *LifecycleDatabase {
		return &LifecycleDatabase{Name: "db"}
	}, di.LifecycleHooks{
		OnStart: func(instance any) error { return instance.(*LifecycleDatabase).Connect() },
		OnStop:  func(instance any) error { return instance.(*LifecycleDatabase).Close() },
	})

	di.SingletonWithLifecycle(func() *LifecycleCache {
		return &LifecycleCache{Name: "cache"}
	}, di.LifecycleHooks{
		OnStart: func(instance any) error { return instance.(*LifecycleCache).Start() },
		OnStop:  func(instance any) error { return instance.(*LifecycleCache).Stop() },
	})

	di.SingletonWithLifecycle(func() *LifecycleServer {
		return &LifecycleServer{Port: 8080}
	}, di.LifecycleHooks{
		OnStart: func(instance any) error { return instance.(*LifecycleServer).Start() },
		OnStop:  func(instance any) error { return instance.(*LifecycleServer).Stop() },
	})

	// Start all
	if err := di.StartAll(); err != nil {
		t.Fatalf("StartAll failed: %v", err)
	}

	// Get instances to verify
	db := di.Resolve[*LifecycleDatabase]()
	cache := di.Resolve[*LifecycleCache]()
	server := di.Resolve[*LifecycleServer]()

	// Verify all started
	if !db.Connected {
		t.Error("Database should be connected")
	}
	if !cache.Started {
		t.Error("Cache should be started")
	}
	if !server.Running {
		t.Error("Server should be running")
	}

	// Stop all (should stop in reverse order: server, cache, db)
	if err := di.StopAll(); err != nil {
		t.Fatalf("StopAll failed: %v", err)
	}

	// Verify all stopped
	if !db.Closed {
		t.Error("Database should be closed")
	}
	if !cache.Stopped {
		t.Error("Cache should be stopped")
	}
	if server.Running {
		t.Error("Server should be stopped")
	}
}

func TestDI_Lifecycle_OnlyOnStart(t *testing.T) {
	di.Reset()

	started := false

	di.SingletonWithLifecycle(func() *LifecycleDatabase {
		return &LifecycleDatabase{Name: "test"}
	}, di.LifecycleHooks{
		OnStart: func(instance any) error {
			started = true
			return nil
		},
		// No OnStop
	})

	if err := di.StartAll(); err != nil {
		t.Fatalf("StartAll failed: %v", err)
	}

	if !started {
		t.Error("OnStart should have been called")
	}

	// StopAll should not fail even with no OnStop
	if err := di.StopAll(); err != nil {
		t.Errorf("StopAll should not fail: %v", err)
	}
}

func TestDI_Lifecycle_OnlyOnStop(t *testing.T) {
	di.Reset()

	stopped := false

	di.SingletonWithLifecycle(func() *LifecycleDatabase {
		return &LifecycleDatabase{Name: "test"}
	}, di.LifecycleHooks{
		// No OnStart
		OnStop: func(instance any) error {
			stopped = true
			return nil
		},
	})

	// Resolve to create instance
	_ = di.Resolve[*LifecycleDatabase]()

	// StartAll should not fail even with no OnStart
	if err := di.StartAll(); err != nil {
		t.Errorf("StartAll should not fail: %v", err)
	}

	if err := di.StopAll(); err != nil {
		t.Fatalf("StopAll failed: %v", err)
	}

	if !stopped {
		t.Error("OnStop should have been called")
	}
}

// --- Error Handling Tests ---

func TestDI_Lifecycle_StartError_Rollback(t *testing.T) {
	di.Reset()

	// Database starts successfully
	di.SingletonWithLifecycle(func() *LifecycleDatabase {
		return &LifecycleDatabase{Name: "db"}
	}, di.LifecycleHooks{
		OnStart: func(instance any) error { return instance.(*LifecycleDatabase).Connect() },
		OnStop:  func(instance any) error { return instance.(*LifecycleDatabase).Close() },
	})

	// Cache fails to start
	di.SingletonWithLifecycle(func() *LifecycleCache {
		return &LifecycleCache{Name: "cache"}
	}, di.LifecycleHooks{
		OnStart: func(instance any) error {
			return errors.New("cache connection failed")
		},
		OnStop: func(instance any) error { return instance.(*LifecycleCache).Stop() },
	})

	// StartAll should fail
	err := di.StartAll()
	if err == nil {
		t.Fatal("StartAll should fail when OnStart returns error")
	}

	if !strings.Contains(err.Error(), "cache connection failed") {
		t.Errorf("Error should mention cache failure: %v", err)
	}

	// Get database to verify rollback
	db := di.Resolve[*LifecycleDatabase]()

	// Database should have been rolled back (closed)
	if !db.Closed {
		t.Error("Database should be closed after rollback")
	}
}

func TestDI_Lifecycle_StopError_ContinuesOthers(t *testing.T) {
	di.Reset()

	di.SingletonWithLifecycle(func() *LifecycleDatabase {
		return &LifecycleDatabase{Name: "db"}
	}, di.LifecycleHooks{
		OnStart: func(instance any) error { return instance.(*LifecycleDatabase).Connect() },
		OnStop: func(instance any) error {
			return errors.New("database close failed")
		},
	})

	di.SingletonWithLifecycle(func() *LifecycleCache {
		return &LifecycleCache{Name: "cache"}
	}, di.LifecycleHooks{
		OnStart: func(instance any) error { return instance.(*LifecycleCache).Start() },
		OnStop:  func(instance any) error { return instance.(*LifecycleCache).Stop() },
	})

	// Start all
	if err := di.StartAll(); err != nil {
		t.Fatalf("StartAll failed: %v", err)
	}

	// Get instances
	cache := di.Resolve[*LifecycleCache]()

	// Stop all - should return error but continue stopping others
	err := di.StopAll()
	if err == nil {
		t.Fatal("StopAll should return error when OnStop fails")
	}

	// Cache should still be stopped even though db stop failed
	if !cache.Stopped {
		t.Error("Cache should be stopped even though database stop failed")
	}
}

// --- Named Lifecycle Tests ---

func TestDI_Lifecycle_Named(t *testing.T) {
	di.Reset()

	di.SingletonNamedWithLifecycle[*LifecycleDatabase]("primary",
		func() *LifecycleDatabase { return &LifecycleDatabase{Name: "primary"} },
		di.LifecycleHooks{
			OnStart: func(instance any) error { return instance.(*LifecycleDatabase).Connect() },
			OnStop:  func(instance any) error { return instance.(*LifecycleDatabase).Close() },
		},
	)

	di.SingletonNamedWithLifecycle[*LifecycleDatabase]("backup",
		func() *LifecycleDatabase { return &LifecycleDatabase{Name: "backup"} },
		di.LifecycleHooks{
			OnStart: func(instance any) error { return instance.(*LifecycleDatabase).Connect() },
			OnStop:  func(instance any) error { return instance.(*LifecycleDatabase).Close() },
		},
	)

	// Start all
	if err := di.StartAll(); err != nil {
		t.Fatalf("StartAll failed: %v", err)
	}

	// Get instances
	primaryDB := di.ResolveNamed[*LifecycleDatabase]("primary")
	backupDB := di.ResolveNamed[*LifecycleDatabase]("backup")

	// Both should be started
	if !primaryDB.Connected {
		t.Error("Primary database should be connected")
	}
	if !backupDB.Connected {
		t.Error("Backup database should be connected")
	}

	// Stop all
	if err := di.StopAll(); err != nil {
		t.Fatalf("StopAll failed: %v", err)
	}

	// Both should be stopped
	if !primaryDB.Closed {
		t.Error("Primary database should be closed")
	}
	if !backupDB.Closed {
		t.Error("Backup database should be closed")
	}
}

// --- Context Tests ---

func TestDI_Lifecycle_WithContext_Timeout(t *testing.T) {
	di.Reset()

	di.SingletonWithLifecycle(func() *LifecycleDatabase {
		return &LifecycleDatabase{Name: "slow"}
	}, di.LifecycleHooks{
		OnStart: func(instance any) error {
			time.Sleep(100 * time.Millisecond)
			return instance.(*LifecycleDatabase).Connect()
		},
	})

	// Context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := di.StartAllWithContext(ctx)
	if err == nil {
		t.Fatal("StartAllWithContext should fail with timeout")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded, got: %v", err)
	}
}

func TestDI_Lifecycle_WithTimeout(t *testing.T) {
	di.Reset()

	di.SingletonWithLifecycle(func() *LifecycleDatabase {
		return &LifecycleDatabase{Name: "fast"}
	}, di.LifecycleHooks{
		OnStart: func(instance any) error {
			return instance.(*LifecycleDatabase).Connect()
		},
	})

	// Should succeed within timeout
	err := di.StartAllWithTimeout(1 * time.Second)
	if err != nil {
		t.Fatalf("StartAllWithTimeout should succeed: %v", err)
	}

	// Get instance to verify
	db := di.Resolve[*LifecycleDatabase]()

	if !db.Connected {
		t.Error("Database should be connected")
	}
}

// --- Edge Cases ---

func TestDI_Lifecycle_NoLifecycleProviders(t *testing.T) {
	di.Reset()

	// Register provider without lifecycle
	di.Singleton(func() *LifecycleDatabase {
		return &LifecycleDatabase{Name: "no-lifecycle"}
	})

	// StartAll and StopAll should not fail
	if err := di.StartAll(); err != nil {
		t.Errorf("StartAll should not fail with no lifecycle providers: %v", err)
	}

	if err := di.StopAll(); err != nil {
		t.Errorf("StopAll should not fail with no lifecycle providers: %v", err)
	}
}

func TestDI_Lifecycle_MultipleStartStop(t *testing.T) {
	di.Reset()

	startCount := 0
	stopCount := 0

	di.SingletonWithLifecycle(func() *LifecycleDatabase {
		return &LifecycleDatabase{Name: "test"}
	}, di.LifecycleHooks{
		OnStart: func(instance any) error {
			startCount++
			return instance.(*LifecycleDatabase).Connect()
		},
		OnStop: func(instance any) error {
			stopCount++
			return instance.(*LifecycleDatabase).Close()
		},
	})

	// First start
	if err := di.StartAll(); err != nil {
		t.Fatalf("First StartAll failed: %v", err)
	}

	// Second start - should skip already started
	if err := di.StartAll(); err != nil {
		t.Fatalf("Second StartAll failed: %v", err)
	}

	// OnStart should only be called once
	if startCount != 1 {
		t.Errorf("OnStart should be called once, was called %d times", startCount)
	}

	// Stop
	if err := di.StopAll(); err != nil {
		t.Fatalf("StopAll failed: %v", err)
	}

	if stopCount != 1 {
		t.Errorf("OnStop should be called once, was called %d times", stopCount)
	}

	// Second stop - should skip not started
	if err := di.StopAll(); err != nil {
		t.Fatalf("Second StopAll failed: %v", err)
	}

	// OnStop should still only be called once
	if stopCount != 1 {
		t.Errorf("OnStop should still be called once, was called %d times", stopCount)
	}
}

func TestDI_Lifecycle_TransientWithLifecycle(t *testing.T) {
	di.Reset()

	// Transient (non-singleton) with lifecycle
	di.RegisterWithLifecycle(func() *LifecycleDatabase {
		return &LifecycleDatabase{Name: "transient"}
	}, di.LifecycleHooks{
		OnStart: func(instance any) error {
			return instance.(*LifecycleDatabase).Connect()
		},
	})

	// Transient providers are not cached, so lifecycle might not work as expected
	// This test documents the behavior
	if err := di.StartAll(); err != nil {
		t.Logf("StartAll with transient lifecycle: %v", err)
	}
}

// --- Integration Tests ---

func TestDI_Lifecycle_RealWorldScenario(t *testing.T) {
	di.Reset()

	// Simulate real application startup
	type Config struct {
		DBHost string
	}

	// Register config (no lifecycle)
	di.Singleton(func() *Config {
		return &Config{DBHost: "localhost:5432"}
	})

	// Register database with lifecycle
	di.SingletonWithLifecycle(func(cfg *Config) *LifecycleDatabase {
		return &LifecycleDatabase{Name: cfg.DBHost}
	}, di.LifecycleHooks{
		OnStart: func(instance any) error {
			return instance.(*LifecycleDatabase).Connect()
		},
		OnStop: func(instance any) error {
			return instance.(*LifecycleDatabase).Close()
		},
	})

	// Register cache with lifecycle
	di.SingletonWithLifecycle(func() *LifecycleCache {
		return &LifecycleCache{Name: "app-cache"}
	}, di.LifecycleHooks{
		OnStart: func(instance any) error {
			return instance.(*LifecycleCache).Start()
		},
		OnStop: func(instance any) error {
			return instance.(*LifecycleCache).Stop()
		},
	})

	// Application startup
	if err := di.StartAll(); err != nil {
		t.Fatalf("Application startup failed: %v", err)
	}

	// Get instances to verify
	db := di.Resolve[*LifecycleDatabase]()
	cache := di.Resolve[*LifecycleCache]()

	// Verify everything is running
	if !db.Connected {
		t.Error("Database should be connected")
	}
	if !cache.Started {
		t.Error("Cache should be started")
	}

	// Application shutdown
	if err := di.StopAll(); err != nil {
		t.Fatalf("Application shutdown failed: %v", err)
	}

	// Verify clean shutdown
	if !db.Closed {
		t.Error("Database should be closed")
	}
	if !cache.Stopped {
		t.Error("Cache should be stopped")
	}
}

func TestDI_Lifecycle_DeferPattern(t *testing.T) {
	di.Reset()

	di.SingletonWithLifecycle(func() *LifecycleDatabase {
		return &LifecycleDatabase{Name: "test"}
	}, di.LifecycleHooks{
		OnStart: func(instance any) error { return instance.(*LifecycleDatabase).Connect() },
		OnStop:  func(instance any) error { return instance.(*LifecycleDatabase).Close() },
	})

	// Simulate main function pattern
	func() {
		if err := di.StartAll(); err != nil {
			t.Fatalf("Startup failed: %v", err)
		}
		defer di.StopAll()

		// Get instance to verify
		db := di.Resolve[*LifecycleDatabase]()

		// Application logic here
		if !db.Connected {
			t.Error("Database should be connected during app logic")
		}
	}()

	// After defer - get instance again
	db := di.Resolve[*LifecycleDatabase]()

	if !db.Closed {
		t.Error("Database should be closed after defer")
	}
}
