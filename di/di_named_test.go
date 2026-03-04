package di_test

import (
	"strings"
	"testing"

	"github.com/leandroluk/gox/di"
)

// --- Types for Named Instance Testing ---

type Database interface {
	Connect() error
	Query(sql string) ([]string, error)
}

type MySQLDatabase struct {
	Host string
}

func (m *MySQLDatabase) Connect() error                     { return nil }
func (m *MySQLDatabase) Query(sql string) ([]string, error) { return []string{"mysql"}, nil }

type PostgresDatabase struct {
	Host string
}

func (p *PostgresDatabase) Connect() error                     { return nil }
func (p *PostgresDatabase) Query(sql string) ([]string, error) { return []string{"postgres"}, nil }

type NamedCache interface {
	Get(key string) string
	Set(key, value string)
}

type NamedRedisCache struct {
	Host string
}

func (r *NamedRedisCache) Get(key string) string { return "redis:" + key }
func (r *NamedRedisCache) Set(key, value string) {}

type MemoryNamedCache struct {
	data map[string]string
}

func (m *MemoryNamedCache) Get(key string) string { return "memory:" + key }
func (m *MemoryNamedCache) Set(key, value string) {}

// --- Named Registration Tests ---

func TestDI_RegisterNamed_Basic(t *testing.T) {
	di.Reset()

	// Register two named databases
	di.RegisterNamed[Database]("mysql", func() *MySQLDatabase {
		return &MySQLDatabase{Host: "localhost:3306"}
	})

	di.RegisterNamed[Database]("postgres", func() *PostgresDatabase {
		return &PostgresDatabase{Host: "localhost:5432"}
	})

	// Resolve by name
	mysql := di.ResolveNamed[Database]("mysql")
	postgres := di.ResolveNamed[Database]("postgres")

	if mysql == nil || postgres == nil {
		t.Fatal("Named providers should resolve successfully")
	}

	// Verify types
	if _, ok := mysql.(*MySQLDatabase); !ok {
		t.Error("mysql should be *MySQLDatabase")
	}

	if _, ok := postgres.(*PostgresDatabase); !ok {
		t.Error("postgres should be *PostgresDatabase")
	}
}

func TestDI_RegisterNamed_EmptyName(t *testing.T) {
	di.Reset()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Should panic when registering with empty name")
		}

		errMsg := r.(string)
		if !strings.Contains(errMsg, "non-empty name") {
			t.Errorf("Error should mention non-empty name, got: %s", errMsg)
		}
	}()

	di.RegisterNamed[Database]("", func() *MySQLDatabase {
		return &MySQLDatabase{}
	})
}

func TestDI_RegisterNamed_DuplicateName(t *testing.T) {
	di.Reset()

	// Register first time - OK
	di.RegisterNamed[Database]("mysql", func() *MySQLDatabase {
		return &MySQLDatabase{Host: "host1"}
	})

	// Register same name again - should panic
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Should panic when registering duplicate name")
		}

		errMsg := r.(string)
		if !strings.Contains(errMsg, "already registered") {
			t.Errorf("Error should mention already registered, got: %s", errMsg)
		}
	}()

	di.RegisterNamed[Database]("mysql", func() *MySQLDatabase {
		return &MySQLDatabase{Host: "host2"}
	})
}

func TestDI_RegisterNamed_UnnamedAndNamed(t *testing.T) {
	di.Reset()

	// Register unnamed (default)
	di.RegisterAs[Database](func() *MySQLDatabase {
		return &MySQLDatabase{Host: "default"}
	})

	// Register named
	di.RegisterNamed[Database]("postgres", func() *PostgresDatabase {
		return &PostgresDatabase{Host: "named"}
	})

	// Both should work
	defaultDB := di.Resolve[Database]()
	namedDB := di.ResolveNamed[Database]("postgres")

	if defaultDB == nil || namedDB == nil {
		t.Fatal("Both unnamed and named should resolve")
	}

	if _, ok := defaultDB.(*MySQLDatabase); !ok {
		t.Error("default should be MySQLDatabase")
	}

	if _, ok := namedDB.(*PostgresDatabase); !ok {
		t.Error("named should be PostgresDatabase")
	}
}

// --- Singleton Named Tests ---

func TestDI_SingletonNamed(t *testing.T) {
	di.Reset()

	di.SingletonNamed[NamedCache]("redis", func() *NamedRedisCache {
		return &NamedRedisCache{Host: "localhost:6379"}
	})

	// Resolve twice
	cache1 := di.ResolveNamed[NamedCache]("redis")
	cache2 := di.ResolveNamed[NamedCache]("redis")

	// Should be same instance
	redis1 := cache1.(*NamedRedisCache)
	redis2 := cache2.(*NamedRedisCache)

	if redis1 != redis2 {
		t.Error("SingletonNamed should return same instance")
	}
}

func TestDI_SingletonInstanceNamed(t *testing.T) {
	di.Reset()

	// Create specific instance
	specificNamedCache := &NamedRedisCache{Host: "production:6379"}
	di.SingletonInstanceNamed[NamedCache]("prod-cache", specificNamedCache)

	// Resolve
	resolved := di.ResolveNamed[NamedCache]("prod-cache")

	// Should be exact same instance
	if resolved.(*NamedRedisCache) != specificNamedCache {
		t.Error("SingletonInstanceNamed should return exact same instance")
	}
}

// --- Resolution Tests ---

func TestDI_ResolveNamed_NotFound(t *testing.T) {
	di.Reset()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("ResolveNamed should panic when name not found")
		}

		errMsg := r.(string)
		if !strings.Contains(errMsg, "nonexistent") {
			t.Errorf("Error should mention the name, got: %s", errMsg)
		}
	}()

	di.ResolveNamed[Database]("nonexistent")
}

func TestDI_ResolveNamed_EmptyName(t *testing.T) {
	di.Reset()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("ResolveNamed should panic with empty name")
		}

		errMsg := r.(string)
		if !strings.Contains(errMsg, "non-empty name") {
			t.Errorf("Error should mention non-empty name, got: %s", errMsg)
		}
	}()

	di.ResolveNamed[Database]("")
}

func TestDI_TryResolveNamed_Success(t *testing.T) {
	di.Reset()

	di.RegisterNamed[NamedCache]("redis", func() *NamedRedisCache {
		return &NamedRedisCache{Host: "localhost"}
	})

	// Should succeed
	if cache, ok := di.TryResolveNamed[NamedCache]("redis"); ok {
		if cache == nil {
			t.Error("NamedCache should not be nil")
		}
	} else {
		t.Error("TryResolveNamed should succeed")
	}
}

func TestDI_TryResolveNamed_NotFound(t *testing.T) {
	di.Reset()

	// Should fail gracefully
	if cache, ok := di.TryResolveNamed[NamedCache]("nonexistent"); ok {
		t.Error("TryResolveNamed should return false for nonexistent name")
	} else {
		if cache != nil {
			t.Error("NamedCache should be nil when not found")
		}
	}
}

func TestDI_TryResolveNamed_EmptyName(t *testing.T) {
	di.Reset()

	// Should fail gracefully with empty name
	if _, ok := di.TryResolveNamed[NamedCache](""); ok {
		t.Error("TryResolveNamed should return false for empty name")
	}
}

// --- MustResolve Named Tests ---

func TestDI_MustResolveNamed_Success(t *testing.T) {
	di.Reset()

	di.RegisterNamed[Database]("mysql", func() *MySQLDatabase {
		return &MySQLDatabase{}
	})

	// Should not panic
	db := di.MustResolveNamed[Database]("mysql", "MySQL is required")

	if db == nil {
		t.Fatal("MustResolveNamed should return instance")
	}
}

func TestDI_MustResolveNamed_CustomMessage(t *testing.T) {
	di.Reset()

	customMsg := "Production database configuration is missing"

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("MustResolveNamed should panic when not found")
		}

		errMsg := r.(string)
		if errMsg != customMsg {
			t.Errorf("Expected custom message %q, got %q", customMsg, errMsg)
		}
	}()

	di.MustResolveNamed[Database]("prod-db", customMsg)
}

// --- ResolveAll Named Tests ---

func TestDI_ResolveAll_WithNamed(t *testing.T) {
	di.Reset()

	// Register unnamed + 2 named
	di.RegisterAs[NamedCache](func() *MemoryNamedCache {
		return &MemoryNamedCache{}
	})

	di.RegisterNamed[NamedCache]("redis", func() *NamedRedisCache {
		return &NamedRedisCache{Host: "host1"}
	})

	di.RegisterNamed[NamedCache]("redis-backup", func() *NamedRedisCache {
		return &NamedRedisCache{Host: "host2"}
	})

	// ResolveAll should return all 3
	caches := di.ResolveAll[NamedCache]()

	if len(caches) != 3 {
		t.Errorf("Expected 3 caches (1 unnamed + 2 named), got %d", len(caches))
	}
}

func TestDI_ResolveAllNamed_OnlyNamed(t *testing.T) {
	di.Reset()

	// Register unnamed + 2 named
	di.RegisterAs[NamedCache](func() *MemoryNamedCache {
		return &MemoryNamedCache{}
	})

	di.RegisterNamed[NamedCache]("redis", func() *NamedRedisCache {
		return &NamedRedisCache{Host: "host1"}
	})

	di.RegisterNamed[NamedCache]("redis-backup", func() *NamedRedisCache {
		return &NamedRedisCache{Host: "host2"}
	})

	// ResolveAllNamed should return only the 2 named (excludes unnamed)
	namedNamedCaches := di.ResolveAllNamed[NamedCache]()

	if len(namedNamedCaches) != 2 {
		t.Errorf("Expected 2 named caches, got %d", len(namedNamedCaches))
	}

	if _, ok := namedNamedCaches["redis"]; !ok {
		t.Error("Should have 'redis' cache")
	}

	if _, ok := namedNamedCaches["redis-backup"]; !ok {
		t.Error("Should have 'redis-backup' cache")
	}
}

func TestDI_ResolveAllNamed_NoNamed(t *testing.T) {
	di.Reset()

	// Register only unnamed
	di.RegisterAs[NamedCache](func() *MemoryNamedCache {
		return &MemoryNamedCache{}
	})

	// ResolveAllNamed should return nil (no named providers)
	namedNamedCaches := di.ResolveAllNamed[NamedCache]()

	if namedNamedCaches != nil {
		t.Error("ResolveAllNamed should return nil when no named providers exist")
	}
}

func TestDI_TryResolveAllNamed_Success(t *testing.T) {
	di.Reset()

	di.RegisterNamed[Database]("mysql", func() *MySQLDatabase { return &MySQLDatabase{} })
	di.RegisterNamed[Database]("postgres", func() *PostgresDatabase { return &PostgresDatabase{} })

	namedDBs, ok := di.TryResolveAllNamed[Database]()

	if !ok {
		t.Fatal("TryResolveAllNamed should succeed when named providers exist")
	}

	if len(namedDBs) != 2 {
		t.Errorf("Expected 2 named databases, got %d", len(namedDBs))
	}
}

func TestDI_TryResolveAllNamed_NotFound(t *testing.T) {
	di.Reset()

	namedDBs, ok := di.TryResolveAllNamed[Database]()

	if ok {
		t.Error("TryResolveAllNamed should return false when no providers exist")
	}

	if namedDBs != nil {
		t.Error("Result should be nil when not found")
	}
}

// --- Real-World Scenarios ---

func TestDI_RealWorld_MultiDatabase(t *testing.T) {
	di.Reset()

	// Setup: application with multiple databases
	di.SingletonNamed[Database]("primary", func() *MySQLDatabase {
		return &MySQLDatabase{Host: "primary.db.local"}
	})

	di.SingletonNamed[Database]("analytics", func() *PostgresDatabase {
		return &PostgresDatabase{Host: "analytics.db.local"}
	})

	di.SingletonNamed[Database]("cache-db", func() *MySQLDatabase {
		return &MySQLDatabase{Host: "cache.db.local"}
	})

	// Usage in app
	primaryDB := di.ResolveNamed[Database]("primary")
	analyticsDB := di.ResolveNamed[Database]("analytics")

	if primaryDB == nil || analyticsDB == nil {
		t.Fatal("Databases should resolve correctly")
	}

	// Verify they're different instances
	if primaryDB == analyticsDB {
		t.Error("Different named instances should be different objects")
	}
}

func TestDI_RealWorld_NamedCacheFallback(t *testing.T) {
	di.Reset()

	// Primary cache (required)
	di.SingletonNamed[NamedCache]("primary", func() *NamedRedisCache {
		return &NamedRedisCache{Host: "redis-primary"}
	})

	// Backup cache (optional)
	// Not registered

	// Usage
	primaryNamedCache := di.ResolveNamed[NamedCache]("primary")

	var backupNamedCache NamedCache
	if cache, ok := di.TryResolveNamed[NamedCache]("backup"); ok {
		backupNamedCache = cache
	}

	if primaryNamedCache == nil {
		t.Fatal("Primary cache should exist")
	}

	if backupNamedCache != nil {
		t.Error("Backup cache should be nil (not registered)")
	}
}

func TestDI_RealWorld_MultiTenant(t *testing.T) {
	di.Reset()

	// Different DB per tenant
	di.RegisterNamed[Database]("tenant-acme", func() *MySQLDatabase {
		return &MySQLDatabase{Host: "acme.db"}
	})

	di.RegisterNamed[Database]("tenant-globex", func() *PostgresDatabase {
		return &PostgresDatabase{Host: "globex.db"}
	})

	// Resolve based on tenant ID
	getTenantDB := func(tenantID string) Database {
		dbName := "tenant-" + tenantID
		return di.ResolveNamed[Database](dbName)
	}

	acmeDB := getTenantDB("acme")
	globexDB := getTenantDB("globex")

	if acmeDB == nil || globexDB == nil {
		t.Fatal("Tenant databases should resolve")
	}
}

func TestDI_RealWorld_FeatureToggle(t *testing.T) {
	di.Reset()

	// Register old implementation (unnamed/default)
	di.RegisterAs[NamedCache](func() *MemoryNamedCache {
		return &MemoryNamedCache{}
	})

	// Register new implementation (named)
	di.RegisterNamed[NamedCache]("new-redis", func() *NamedRedisCache {
		return &NamedRedisCache{Host: "new-feature"}
	})

	// Feature flag logic
	useNewNamedCache := true

	var cache NamedCache
	if useNewNamedCache {
		if c, ok := di.TryResolveNamed[NamedCache]("new-redis"); ok {
			cache = c
		} else {
			cache = di.Resolve[NamedCache]() // Fallback to default
		}
	} else {
		cache = di.Resolve[NamedCache]()
	}

	if cache == nil {
		t.Fatal("NamedCache should resolve")
	}

	// With feature enabled, should get Redis
	if _, ok := cache.(*NamedRedisCache); !ok {
		t.Error("Should have resolved to NamedRedisCache with feature enabled")
	}
}

// --- Edge Cases ---

func TestDI_Named_WithDependencies(t *testing.T) {
	di.Reset()

	type Config struct {
		Name string
	}

	type Service struct {
		DB     Database
		Config *Config
	}

	// Register unnamed dependencies
	di.Register(func() *Config {
		return &Config{Name: "default"}
	})

	// Register named service with unnamed dependencies
	di.RegisterNamed[*Service]("api-service", func(db Database, cfg *Config) *Service {
		return &Service{DB: db, Config: cfg}
	})

	// This should fail because Database dependency is not registered
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Should panic when unnamed dependency is missing")
		}
	}()

	di.ResolveNamed[*Service]("api-service")
}

func TestDI_Named_TransientVsSingleton(t *testing.T) {
	di.Reset()

	// Transient named
	di.RegisterNamed[NamedCache]("transient", func() *MemoryNamedCache {
		return &MemoryNamedCache{}
	})

	// Singleton named
	di.SingletonNamed[NamedCache]("singleton", func() *NamedRedisCache {
		return &NamedRedisCache{}
	})

	// Transient should create new instances
	trans1 := di.ResolveNamed[NamedCache]("transient").(*MemoryNamedCache)
	trans2 := di.ResolveNamed[NamedCache]("transient").(*MemoryNamedCache)

	if trans1 == trans2 {
		t.Error("Transient named should create different instances")
	}

	// Singleton should return same instance
	sing1 := di.ResolveNamed[NamedCache]("singleton").(*NamedRedisCache)
	sing2 := di.ResolveNamed[NamedCache]("singleton").(*NamedRedisCache)

	if sing1 != sing2 {
		t.Error("Singleton named should return same instance")
	}
}

func TestDI_Named_ResetClearsAll(t *testing.T) {
	di.Reset()

	// Register various providers
	di.Register(func() *MySQLDatabase { return &MySQLDatabase{} })
	di.RegisterNamed[Database]("mysql", func() *MySQLDatabase { return &MySQLDatabase{} })
	di.RegisterNamed[Database]("postgres", func() *PostgresDatabase { return &PostgresDatabase{} })

	// Reset
	di.Reset()

	// Verify registry is empty
	di.RegistryMutex.RLock()
	count := len(di.ProviderRegistry)
	di.RegistryMutex.RUnlock()

	if count != 0 {
		t.Errorf("Reset should clear all providers, found %d", count)
	}
}
