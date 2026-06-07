package di_v1_test

import (
	"strings"
	"testing"

	"github.com/leandroluk/gox/di"
)

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

func TestDI_RegisterNamed_Basic(t *testing.T) {
	di.Reset()

	di.RegisterNamed[Database]("mysql", func(o *di.Options[Database]) {
		o.Constructor = func() (Database, error) { return &MySQLDatabase{Host: "localhost:3306"}, nil }
	})

	di.RegisterNamed[Database]("postgres", func(o *di.Options[Database]) {
		o.Constructor = func() (Database, error) { return &PostgresDatabase{Host: "localhost:5432"}, nil }
	})

	mysql := di.ResolveNamed[Database]("mysql")
	postgres := di.ResolveNamed[Database]("postgres")

	if mysql == nil || postgres == nil {
		t.Fatal("Named providers should resolve successfully")
	}

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

	di.RegisterNamed[Database]("", func(o *di.Options[Database]) {
		o.Constructor = func() (Database, error) { return &MySQLDatabase{}, nil }
	})
}

func TestDI_RegisterNamed_DuplicateName(t *testing.T) {
	di.Reset()

	di.RegisterNamed[Database]("mysql", func(o *di.Options[Database]) {
		o.Constructor = func() (Database, error) { return &MySQLDatabase{Host: "host1"}, nil }
	})

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

	di.RegisterNamed[Database]("mysql", func(o *di.Options[Database]) {
		o.Constructor = func() (Database, error) { return &MySQLDatabase{Host: "host2"}, nil }
	})
}

func TestDI_RegisterNamed_UnnamedAndNamed(t *testing.T) {
	di.Reset()

	di.RegisterAs[Database](func(o *di.Options[Database]) {
		o.Constructor = func() (Database, error) { return &MySQLDatabase{Host: "default"}, nil }
	})

	di.RegisterNamed[Database]("postgres", func(o *di.Options[Database]) {
		o.Constructor = func() (Database, error) { return &PostgresDatabase{Host: "named"}, nil }
	})

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

func TestDI_SingletonNamed(t *testing.T) {
	di.Reset()

	di.SingletonNamed[NamedCache]("redis", func(o *di.Options[NamedCache]) {
		o.Constructor = func() (NamedCache, error) { return &NamedRedisCache{Host: "localhost:6379"}, nil }
	})

	cache1 := di.ResolveNamed[NamedCache]("redis")
	cache2 := di.ResolveNamed[NamedCache]("redis")

	redis1 := cache1.(*NamedRedisCache)
	redis2 := cache2.(*NamedRedisCache)

	if redis1 != redis2 {
		t.Error("SingletonNamed should return same instance")
	}
}

func TestDI_SingletonInstanceNamed(t *testing.T) {
	di.Reset()

	specificNamedCache := &NamedRedisCache{Host: "production:6379"}
	di.SingletonInstanceNamed[NamedCache]("prod-cache", specificNamedCache, nil)

	resolved := di.ResolveNamed[NamedCache]("prod-cache")

	if resolved.(*NamedRedisCache) != specificNamedCache {
		t.Error("SingletonInstanceNamed should return exact same instance")
	}
}

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

	di.RegisterNamed[NamedCache]("redis", func(o *di.Options[NamedCache]) {
		o.Constructor = func() (NamedCache, error) { return &NamedRedisCache{Host: "localhost"}, nil }
	})

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

	if _, ok := di.TryResolveNamed[NamedCache](""); ok {
		t.Error("TryResolveNamed should return false for empty name")
	}
}

func TestDI_MustResolveNamed_Success(t *testing.T) {
	di.Reset()

	di.RegisterNamed[Database]("mysql", func(o *di.Options[Database]) {
		o.Constructor = func() (Database, error) { return &MySQLDatabase{}, nil }
	})

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

func TestDI_ResolveAll_WithNamed(t *testing.T) {
	di.Reset()

	di.RegisterAs[NamedCache](func(o *di.Options[NamedCache]) {
		o.Constructor = func() (NamedCache, error) { return &MemoryNamedCache{}, nil }
	})

	di.RegisterNamed[NamedCache]("redis", func(o *di.Options[NamedCache]) {
		o.Constructor = func() (NamedCache, error) { return &NamedRedisCache{Host: "host1"}, nil }
	})

	di.RegisterNamed[NamedCache]("redis-backup", func(o *di.Options[NamedCache]) {
		o.Constructor = func() (NamedCache, error) { return &NamedRedisCache{Host: "host2"}, nil }
	})

	caches := di.ResolveAll[NamedCache]()

	if len(caches) != 3 {
		t.Errorf("Expected 3 caches (1 unnamed + 2 named), got %d", len(caches))
	}
}

func TestDI_ResolveAllNamed_OnlyNamed(t *testing.T) {
	di.Reset()

	di.RegisterAs[NamedCache](func(o *di.Options[NamedCache]) {
		o.Constructor = func() (NamedCache, error) { return &MemoryNamedCache{}, nil }
	})

	di.RegisterNamed[NamedCache]("redis", func(o *di.Options[NamedCache]) {
		o.Constructor = func() (NamedCache, error) { return &NamedRedisCache{Host: "host1"}, nil }
	})

	di.RegisterNamed[NamedCache]("redis-backup", func(o *di.Options[NamedCache]) {
		o.Constructor = func() (NamedCache, error) { return &NamedRedisCache{Host: "host2"}, nil }
	})

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

	di.RegisterAs[NamedCache](func(o *di.Options[NamedCache]) {
		o.Constructor = func() (NamedCache, error) { return &MemoryNamedCache{}, nil }
	})

	namedNamedCaches := di.ResolveAllNamed[NamedCache]()

	if namedNamedCaches != nil {
		t.Error("ResolveAllNamed should return nil when no named providers exist")
	}
}

func TestDI_TryResolveAllNamed_Success(t *testing.T) {
	di.Reset()

	di.RegisterNamed[Database]("mysql", func(o *di.Options[Database]) {
		o.Constructor = func() (Database, error) { return &MySQLDatabase{}, nil }
	})
	di.RegisterNamed[Database]("postgres", func(o *di.Options[Database]) {
		o.Constructor = func() (Database, error) { return &PostgresDatabase{}, nil }
	})

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

func TestDI_RealWorld_MultiDatabase(t *testing.T) {
	di.Reset()

	di.SingletonNamed[Database]("primary", func(o *di.Options[Database]) {
		o.Constructor = func() (Database, error) { return &MySQLDatabase{Host: "primary.db.local"}, nil }
	})

	di.SingletonNamed[Database]("analytics", func(o *di.Options[Database]) {
		o.Constructor = func() (Database, error) { return &PostgresDatabase{Host: "analytics.db.local"}, nil }
	})

	di.SingletonNamed[Database]("cache-db", func(o *di.Options[Database]) {
		o.Constructor = func() (Database, error) { return &MySQLDatabase{Host: "cache.db.local"}, nil }
	})

	primaryDB := di.ResolveNamed[Database]("primary")
	analyticsDB := di.ResolveNamed[Database]("analytics")

	if primaryDB == nil || analyticsDB == nil {
		t.Fatal("Databases should resolve correctly")
	}

	if primaryDB == analyticsDB {
		t.Error("Different named instances should be different objects")
	}
}

func TestDI_RealWorld_NamedCacheFallback(t *testing.T) {
	di.Reset()

	di.SingletonNamed[NamedCache]("primary", func(o *di.Options[NamedCache]) {
		o.Constructor = func() (NamedCache, error) { return &NamedRedisCache{Host: "redis-primary"}, nil }
	})

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

	di.RegisterNamed[Database]("tenant-acme", func(o *di.Options[Database]) {
		o.Constructor = func() (Database, error) { return &MySQLDatabase{Host: "acme.db"}, nil }
	})

	di.RegisterNamed[Database]("tenant-globex", func(o *di.Options[Database]) {
		o.Constructor = func() (Database, error) { return &PostgresDatabase{Host: "globex.db"}, nil }
	})

	getTenantDB := func(tenantID string) Database {
		return di.ResolveNamed[Database]("tenant-" + tenantID)
	}

	acmeDB := getTenantDB("acme")
	globexDB := getTenantDB("globex")

	if acmeDB == nil || globexDB == nil {
		t.Fatal("Tenant databases should resolve")
	}
}

func TestDI_RealWorld_FeatureToggle(t *testing.T) {
	di.Reset()

	di.RegisterAs[NamedCache](func(o *di.Options[NamedCache]) {
		o.Constructor = func() (NamedCache, error) { return &MemoryNamedCache{}, nil }
	})

	di.RegisterNamed[NamedCache]("new-redis", func(o *di.Options[NamedCache]) {
		o.Constructor = func() (NamedCache, error) { return &NamedRedisCache{Host: "new-feature"}, nil }
	})

	useNewNamedCache := true

	var cache NamedCache
	if useNewNamedCache {
		if c, ok := di.TryResolveNamed[NamedCache]("new-redis"); ok {
			cache = c
		} else {
			cache = di.Resolve[NamedCache]()
		}
	} else {
		cache = di.Resolve[NamedCache]()
	}

	if cache == nil {
		t.Fatal("NamedCache should resolve")
	}

	if _, ok := cache.(*NamedRedisCache); !ok {
		t.Error("Should have resolved to NamedRedisCache with feature enabled")
	}
}

func TestDI_Named_WithDependencies(t *testing.T) {
	di.Reset()

	type Config struct {
		Name string
	}

	type Service struct {
		DB     Database
		Config *Config
	}

	di.Register[*Config](func(o *di.Options[*Config]) {
		o.Constructor = func() (*Config, error) { return &Config{Name: "default"}, nil }
	})

	di.RegisterNamed[*Service]("api-service", func(o *di.Options[*Service]) {
		o.Constructor = func() (*Service, error) {
			db := di.Resolve[Database]()
			cfg := di.Resolve[*Config]()
			return &Service{DB: db, Config: cfg}, nil
		}
	})

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Should panic when unnamed dependency is missing")
		}
	}()

	di.ResolveNamed[*Service]("api-service")
}

func TestDI_Named_TransientVsSingleton(t *testing.T) {
	di.Reset()

	di.RegisterNamed[NamedCache]("transient", func(o *di.Options[NamedCache]) {
		o.Constructor = func() (NamedCache, error) { return &MemoryNamedCache{}, nil }
	})

	di.SingletonNamed[NamedCache]("singleton", func(o *di.Options[NamedCache]) {
		o.Constructor = func() (NamedCache, error) { return &NamedRedisCache{}, nil }
	})

	trans1 := di.ResolveNamed[NamedCache]("transient").(*MemoryNamedCache)
	trans2 := di.ResolveNamed[NamedCache]("transient").(*MemoryNamedCache)

	if trans1 == trans2 {
		t.Error("Transient named should create different instances")
	}

	sing1 := di.ResolveNamed[NamedCache]("singleton").(*NamedRedisCache)
	sing2 := di.ResolveNamed[NamedCache]("singleton").(*NamedRedisCache)

	if sing1 != sing2 {
		t.Error("Singleton named should return same instance")
	}
}

func TestDI_Named_ResetClearsAll(t *testing.T) {
	di.Reset()

	di.Register[*MySQLDatabase](func(o *di.Options[*MySQLDatabase]) {
		o.Constructor = func() (*MySQLDatabase, error) { return &MySQLDatabase{}, nil }
	})
	di.RegisterNamed[Database]("mysql", func(o *di.Options[Database]) {
		o.Constructor = func() (Database, error) { return &MySQLDatabase{}, nil }
	})
	di.RegisterNamed[Database]("postgres", func(o *di.Options[Database]) {
		o.Constructor = func() (Database, error) { return &PostgresDatabase{}, nil }
	})

	di.Reset()

	di.RegistryMutex.RLock()
	count := len(di.ProviderRegistry)
	di.RegistryMutex.RUnlock()

	if count != 0 {
		t.Errorf("Reset should clear all providers, found %d", count)
	}
}
