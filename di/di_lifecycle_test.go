package di_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/leandroluk/gox/di"
)

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

func TestDI_Lifecycle_BasicStartStop(t *testing.T) {
	di.Reset()

	di.Singleton[*LifecycleDatabase](func(o *di.Options[*LifecycleDatabase]) {
		o.Constructor = func() (*LifecycleDatabase, error) {
			return &LifecycleDatabase{Name: "test"}, nil
		}
		o.OnApplicationBootstrap = func(db *LifecycleDatabase) error {
			return db.Connect()
		}
		o.OnApplicationShutdown = func(db *LifecycleDatabase) error {
			return db.Close()
		}
	})

	if err := di.StartAll(); err != nil {
		t.Fatalf("StartAll failed: %v", err)
	}

	db := di.Resolve[*LifecycleDatabase]()

	if !db.Connected {
		t.Error("Database should be connected after StartAll")
	}

	if err := di.StopAll(); err != nil {
		t.Fatalf("StopAll failed: %v", err)
	}

	if !db.Closed {
		t.Error("Database should be closed after StopAll")
	}
}

func TestDI_Lifecycle_MultipleProviders(t *testing.T) {
	di.Reset()

	di.Singleton[*LifecycleDatabase](func(o *di.Options[*LifecycleDatabase]) {
		o.Constructor = func() (*LifecycleDatabase, error) { return &LifecycleDatabase{Name: "db"}, nil }
		o.OnApplicationBootstrap = func(db *LifecycleDatabase) error { return db.Connect() }
		o.OnApplicationShutdown = func(db *LifecycleDatabase) error { return db.Close() }
	})

	di.Singleton[*LifecycleCache](func(o *di.Options[*LifecycleCache]) {
		o.Constructor = func() (*LifecycleCache, error) { return &LifecycleCache{Name: "cache"}, nil }
		o.OnApplicationBootstrap = func(c *LifecycleCache) error { return c.Start() }
		o.OnApplicationShutdown = func(c *LifecycleCache) error { return c.Stop() }
	})

	di.Singleton[*LifecycleServer](func(o *di.Options[*LifecycleServer]) {
		o.Constructor = func() (*LifecycleServer, error) { return &LifecycleServer{Port: 8080}, nil }
		o.OnApplicationBootstrap = func(s *LifecycleServer) error { return s.Start() }
		o.OnApplicationShutdown = func(s *LifecycleServer) error { return s.Stop() }
	})

	if err := di.StartAll(); err != nil {
		t.Fatalf("StartAll failed: %v", err)
	}

	db := di.Resolve[*LifecycleDatabase]()
	cache := di.Resolve[*LifecycleCache]()
	server := di.Resolve[*LifecycleServer]()

	if !db.Connected {
		t.Error("Database should be connected")
	}
	if !cache.Started {
		t.Error("Cache should be started")
	}
	if !server.Running {
		t.Error("Server should be running")
	}

	if err := di.StopAll(); err != nil {
		t.Fatalf("StopAll failed: %v", err)
	}

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

func TestDI_Lifecycle_OnlyOnBootstrap(t *testing.T) {
	di.Reset()

	started := false

	di.Singleton[*LifecycleDatabase](func(o *di.Options[*LifecycleDatabase]) {
		o.Constructor = func() (*LifecycleDatabase, error) { return &LifecycleDatabase{Name: "test"}, nil }
		o.OnApplicationBootstrap = func(_ *LifecycleDatabase) error {
			started = true
			return nil
		}
	})

	if err := di.StartAll(); err != nil {
		t.Fatalf("StartAll failed: %v", err)
	}

	if !started {
		t.Error("OnApplicationBootstrap should have been called")
	}

	if err := di.StopAll(); err != nil {
		t.Errorf("StopAll should not fail: %v", err)
	}
}

func TestDI_Lifecycle_OnlyOnShutdown(t *testing.T) {
	di.Reset()

	stopped := false

	di.Singleton[*LifecycleDatabase](func(o *di.Options[*LifecycleDatabase]) {
		o.Constructor = func() (*LifecycleDatabase, error) { return &LifecycleDatabase{Name: "test"}, nil }
		o.OnApplicationShutdown = func(_ *LifecycleDatabase) error {
			stopped = true
			return nil
		}
	})

	_ = di.Resolve[*LifecycleDatabase]()

	if err := di.StartAll(); err != nil {
		t.Errorf("StartAll should not fail: %v", err)
	}

	if err := di.StopAll(); err != nil {
		t.Fatalf("StopAll failed: %v", err)
	}

	if !stopped {
		t.Error("OnApplicationShutdown should have been called")
	}
}

func TestDI_Lifecycle_BootstrapError_Rollback(t *testing.T) {
	di.Reset()

	di.Singleton[*LifecycleDatabase](func(o *di.Options[*LifecycleDatabase]) {
		o.Constructor = func() (*LifecycleDatabase, error) { return &LifecycleDatabase{Name: "db"}, nil }
		o.OnApplicationBootstrap = func(db *LifecycleDatabase) error { return db.Connect() }
		o.OnApplicationShutdown = func(db *LifecycleDatabase) error { return db.Close() }
	})

	di.Singleton[*LifecycleCache](func(o *di.Options[*LifecycleCache]) {
		o.Constructor = func() (*LifecycleCache, error) { return &LifecycleCache{Name: "cache"}, nil }
		o.OnApplicationBootstrap = func(_ *LifecycleCache) error {
			return errors.New("cache connection failed")
		}
		o.OnApplicationShutdown = func(c *LifecycleCache) error { return c.Stop() }
	})

	err := di.StartAll()
	if err == nil {
		t.Fatal("StartAll should fail when OnApplicationBootstrap returns error")
	}

	if !strings.Contains(err.Error(), "cache connection failed") {
		t.Errorf("Error should mention cache failure: %v", err)
	}

	db := di.Resolve[*LifecycleDatabase]()

	if !db.Closed {
		t.Error("Database should be closed after rollback")
	}
}

func TestDI_Lifecycle_ShutdownError_ContinuesOthers(t *testing.T) {
	di.Reset()

	di.Singleton[*LifecycleDatabase](func(o *di.Options[*LifecycleDatabase]) {
		o.Constructor = func() (*LifecycleDatabase, error) { return &LifecycleDatabase{Name: "db"}, nil }
		o.OnApplicationBootstrap = func(db *LifecycleDatabase) error { return db.Connect() }
		o.OnApplicationShutdown = func(_ *LifecycleDatabase) error {
			return errors.New("database close failed")
		}
	})

	di.Singleton[*LifecycleCache](func(o *di.Options[*LifecycleCache]) {
		o.Constructor = func() (*LifecycleCache, error) { return &LifecycleCache{Name: "cache"}, nil }
		o.OnApplicationBootstrap = func(c *LifecycleCache) error { return c.Start() }
		o.OnApplicationShutdown = func(c *LifecycleCache) error { return c.Stop() }
	})

	if err := di.StartAll(); err != nil {
		t.Fatalf("StartAll failed: %v", err)
	}

	cache := di.Resolve[*LifecycleCache]()

	err := di.StopAll()
	if err == nil {
		t.Fatal("StopAll should return error when OnApplicationShutdown fails")
	}

	if !cache.Stopped {
		t.Error("Cache should be stopped even though database stop failed")
	}
}

func TestDI_Lifecycle_Named(t *testing.T) {
	di.Reset()

	di.SingletonNamed[*LifecycleDatabase]("primary", func(o *di.Options[*LifecycleDatabase]) {
		o.Constructor = func() (*LifecycleDatabase, error) { return &LifecycleDatabase{Name: "primary"}, nil }
		o.OnApplicationBootstrap = func(db *LifecycleDatabase) error { return db.Connect() }
		o.OnApplicationShutdown = func(db *LifecycleDatabase) error { return db.Close() }
	})

	di.SingletonNamed[*LifecycleDatabase]("backup", func(o *di.Options[*LifecycleDatabase]) {
		o.Constructor = func() (*LifecycleDatabase, error) { return &LifecycleDatabase{Name: "backup"}, nil }
		o.OnApplicationBootstrap = func(db *LifecycleDatabase) error { return db.Connect() }
		o.OnApplicationShutdown = func(db *LifecycleDatabase) error { return db.Close() }
	})

	if err := di.StartAll(); err != nil {
		t.Fatalf("StartAll failed: %v", err)
	}

	primaryDB := di.ResolveNamed[*LifecycleDatabase]("primary")
	backupDB := di.ResolveNamed[*LifecycleDatabase]("backup")

	if !primaryDB.Connected {
		t.Error("Primary database should be connected")
	}
	if !backupDB.Connected {
		t.Error("Backup database should be connected")
	}

	if err := di.StopAll(); err != nil {
		t.Fatalf("StopAll failed: %v", err)
	}

	if !primaryDB.Closed {
		t.Error("Primary database should be closed")
	}
	if !backupDB.Closed {
		t.Error("Backup database should be closed")
	}
}

func TestDI_Lifecycle_WithContext_Timeout(t *testing.T) {
	di.Reset()

	di.Singleton[*LifecycleDatabase](func(o *di.Options[*LifecycleDatabase]) {
		o.Constructor = func() (*LifecycleDatabase, error) { return &LifecycleDatabase{Name: "slow"}, nil }
		o.OnApplicationBootstrap = func(db *LifecycleDatabase) error {
			time.Sleep(100 * time.Millisecond)
			return db.Connect()
		}
	})

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

	di.Singleton[*LifecycleDatabase](func(o *di.Options[*LifecycleDatabase]) {
		o.Constructor = func() (*LifecycleDatabase, error) { return &LifecycleDatabase{Name: "fast"}, nil }
		o.OnApplicationBootstrap = func(db *LifecycleDatabase) error { return db.Connect() }
	})

	err := di.StartAllWithTimeout(1 * time.Second)
	if err != nil {
		t.Fatalf("StartAllWithTimeout should succeed: %v", err)
	}

	db := di.Resolve[*LifecycleDatabase]()

	if !db.Connected {
		t.Error("Database should be connected")
	}
}

func TestDI_Lifecycle_NoLifecycleProviders(t *testing.T) {
	di.Reset()

	di.Singleton[*LifecycleDatabase](func(o *di.Options[*LifecycleDatabase]) {
		o.Constructor = func() (*LifecycleDatabase, error) { return &LifecycleDatabase{Name: "no-lifecycle"}, nil }
	})

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

	di.Singleton[*LifecycleDatabase](func(o *di.Options[*LifecycleDatabase]) {
		o.Constructor = func() (*LifecycleDatabase, error) { return &LifecycleDatabase{Name: "test"}, nil }
		o.OnApplicationBootstrap = func(db *LifecycleDatabase) error {
			startCount++
			return db.Connect()
		}
		o.OnApplicationShutdown = func(db *LifecycleDatabase) error {
			stopCount++
			return db.Close()
		}
	})

	if err := di.StartAll(); err != nil {
		t.Fatalf("First StartAll failed: %v", err)
	}

	if err := di.StartAll(); err != nil {
		t.Fatalf("Second StartAll failed: %v", err)
	}

	if startCount != 1 {
		t.Errorf("OnApplicationBootstrap should be called once, was called %d times", startCount)
	}

	if err := di.StopAll(); err != nil {
		t.Fatalf("StopAll failed: %v", err)
	}

	if stopCount != 1 {
		t.Errorf("OnApplicationShutdown should be called once, was called %d times", stopCount)
	}

	if err := di.StopAll(); err != nil {
		t.Fatalf("Second StopAll failed: %v", err)
	}

	if stopCount != 1 {
		t.Errorf("OnApplicationShutdown should still be called once, was called %d times", stopCount)
	}
}

func TestDI_Lifecycle_TransientWithLifecycle(t *testing.T) {
	di.Reset()

	di.Register[*LifecycleDatabase](func(o *di.Options[*LifecycleDatabase]) {
		o.Constructor = func() (*LifecycleDatabase, error) { return &LifecycleDatabase{Name: "transient"}, nil }
		o.OnApplicationBootstrap = func(db *LifecycleDatabase) error { return db.Connect() }
	})

	if err := di.StartAll(); err != nil {
		t.Logf("StartAll with transient lifecycle: %v", err)
	}
}

func TestDI_Lifecycle_RealWorldScenario(t *testing.T) {
	di.Reset()

	type Config struct {
		DBHost string
	}

	di.Singleton[*Config](func(o *di.Options[*Config]) {
		o.Constructor = func() (*Config, error) { return &Config{DBHost: "localhost:5432"}, nil }
	})

	di.SingletonFrom[*LifecycleDatabase](func() (*LifecycleDatabase, error) {
		cfg := di.Resolve[*Config]()
		return &LifecycleDatabase{Name: cfg.DBHost}, nil
	})

	// Registra bootstrap/shutdown via Singleton com lifecycle separado
	di.Singleton[*LifecycleCache](func(o *di.Options[*LifecycleCache]) {
		o.Constructor = func() (*LifecycleCache, error) { return &LifecycleCache{Name: "app-cache"}, nil }
		o.OnApplicationBootstrap = func(c *LifecycleCache) error { return c.Start() }
		o.OnApplicationShutdown = func(c *LifecycleCache) error { return c.Stop() }
	})

	if err := di.StartAll(); err != nil {
		t.Fatalf("Application startup failed: %v", err)
	}

	cache := di.Resolve[*LifecycleCache]()

	if !cache.Started {
		t.Error("Cache should be started")
	}

	if err := di.StopAll(); err != nil {
		t.Fatalf("Application shutdown failed: %v", err)
	}

	if !cache.Stopped {
		t.Error("Cache should be stopped")
	}
}

func TestDI_Lifecycle_DeferPattern(t *testing.T) {
	di.Reset()

	di.Singleton[*LifecycleDatabase](func(o *di.Options[*LifecycleDatabase]) {
		o.Constructor = func() (*LifecycleDatabase, error) { return &LifecycleDatabase{Name: "test"}, nil }
		o.OnApplicationBootstrap = func(db *LifecycleDatabase) error { return db.Connect() }
		o.OnApplicationShutdown = func(db *LifecycleDatabase) error { return db.Close() }
	})

	func() {
		if err := di.StartAll(); err != nil {
			t.Fatalf("Startup failed: %v", err)
		}
		defer di.StopAll()

		db := di.Resolve[*LifecycleDatabase]()

		if !db.Connected {
			t.Error("Database should be connected during app logic")
		}
	}()

	db := di.Resolve[*LifecycleDatabase]()

	if !db.Closed {
		t.Error("Database should be closed after defer")
	}
}
