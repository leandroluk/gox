package di_test

import (
	"strings"
	"testing"

	"github.com/leandroluk/gox/di"
)

// --- Circular Dependency Test Types ---

type CircularServiceA struct {
	B *CircularServiceB
}

type CircularServiceB struct {
	C *CircularServiceC
}

type CircularServiceC struct {
	A *CircularServiceA // Circular!
}

type CircularDirectA struct {
	B *CircularDirectB
}

type CircularDirectB struct {
	A *CircularDirectA // Direct circular!
}

type CircularSelfRef struct {
	Self *CircularSelfRef
}

// --- Circular Dependency Tests ---

func TestDI_CircularDependency_Direct(t *testing.T) {
	di.Reset()

	// Register circular dependencies
	di.Register(func(b *CircularDirectB) *CircularDirectA {
		return &CircularDirectA{B: b}
	})
	di.Register(func(a *CircularDirectA) *CircularDirectB {
		return &CircularDirectB{A: a}
	})

	// Should panic with circular dependency message
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Expected panic for circular dependency, got none")
		}

		errMsg := r.(string)

		// Check for key elements in error message
		if !strings.Contains(errMsg, "circular dependency detected") {
			t.Errorf("Error should mention 'circular dependency', got: %s", errMsg)
		}

		if !strings.Contains(errMsg, "CircularDirectA") || !strings.Contains(errMsg, "CircularDirectB") {
			t.Errorf("Error should show types involved in cycle, got: %s", errMsg)
		}

		if !strings.Contains(errMsg, "dependency chain") {
			t.Errorf("Error should show dependency chain, got: %s", errMsg)
		}

		if !strings.Contains(errMsg, "hint:") {
			t.Errorf("Error should include a hint, got: %s", errMsg)
		}
	}()

	// This should trigger the circular dependency
	di.Resolve[*CircularDirectA]()
}

func TestDI_CircularDependency_Indirect(t *testing.T) {
	di.Reset()

	// Register indirect circular dependencies: A -> B -> C -> A
	di.Register(func(b *CircularServiceB) *CircularServiceA {
		return &CircularServiceA{B: b}
	})
	di.Register(func(c *CircularServiceC) *CircularServiceB {
		return &CircularServiceB{C: c}
	})
	di.Register(func(a *CircularServiceA) *CircularServiceC {
		return &CircularServiceC{A: a}
	})

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Expected panic for circular dependency, got none")
		}

		errMsg := r.(string)

		if !strings.Contains(errMsg, "circular dependency detected") {
			t.Errorf("Error should mention 'circular dependency', got: %s", errMsg)
		}

		// Should show at least some of the types in the cycle
		hasServiceType := strings.Contains(errMsg, "CircularServiceA") ||
			strings.Contains(errMsg, "CircularServiceB") ||
			strings.Contains(errMsg, "CircularServiceC")

		if !hasServiceType {
			t.Errorf("Error should show types involved in cycle, got: %s", errMsg)
		}
	}()

	// This should trigger the circular dependency
	di.Resolve[*CircularServiceA]()
}

func TestDI_CircularDependency_WithSingleton(t *testing.T) {
	di.Reset()

	// Even with Singleton, circular dependencies should be detected
	di.Singleton(func(b *CircularDirectB) *CircularDirectA {
		return &CircularDirectA{B: b}
	})
	di.Singleton(func(a *CircularDirectA) *CircularDirectB {
		return &CircularDirectB{A: a}
	})

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Expected panic for circular dependency with singleton, got none")
		}

		errMsg := r.(string)
		if !strings.Contains(errMsg, "circular dependency") {
			t.Errorf("Should detect circular dependency even with Singleton, got: %s", errMsg)
		}
	}()

	di.Resolve[*CircularDirectA]()
}

func TestDI_CircularDependency_SelfReference(t *testing.T) {
	di.Reset()

	// Register a type that depends on itself
	di.Register(func(self *CircularSelfRef) *CircularSelfRef {
		return &CircularSelfRef{Self: self}
	})

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Expected panic for self-referencing dependency, got none")
		}

		errMsg := r.(string)
		if !strings.Contains(errMsg, "circular dependency") {
			t.Errorf("Should detect self-referencing as circular, got: %s", errMsg)
		}
	}()

	di.Resolve[*CircularSelfRef]()
}

func TestDI_NoCircularDependency_ValidChain(t *testing.T) {
	di.Reset()

	type CircularConfig struct {
		Value string
	}
	type CircularDatabase struct {
		Config *CircularConfig
	}
	type CircularService struct {
		DB *CircularDatabase
	}

	// Register valid dependency chain: Service -> Database -> Config
	di.Register(func() *CircularConfig {
		return &CircularConfig{Value: "test"}
	})
	di.Register(func(cfg *CircularConfig) *CircularDatabase {
		return &CircularDatabase{Config: cfg}
	})
	di.Register(func(db *CircularDatabase) *CircularService {
		return &CircularService{DB: db}
	})

	// Should NOT panic - this is a valid chain
	service := di.Resolve[*CircularService]()

	if service == nil {
		t.Fatal("Expected valid service, got nil")
	}

	if service.DB == nil {
		t.Fatal("Expected valid database, got nil")
	}

	if service.DB.Config == nil {
		t.Fatal("Expected valid config, got nil")
	}

	if service.DB.Config.Value != "test" {
		t.Errorf("Expected config value 'test', got '%s'", service.DB.Config.Value)
	}
}

func TestDI_ConcurrentResolution_NoCircular(t *testing.T) {
	di.Reset()

	type CircularSharedService struct {
		ID int
	}

	counter := 0
	di.Register(func() *CircularSharedService {
		counter++
		return &CircularSharedService{ID: counter}
	})

	// Multiple goroutines resolving simultaneously
	const goroutines = 10
	done := make(chan bool, goroutines)
	errors := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					// Check if it's a circular dependency error
					if errMsg, ok := r.(string); ok && strings.Contains(errMsg, "circular") {
						errors <- nil // This would be unexpected
					}
				}
				done <- true
			}()

			service := di.Resolve[*CircularSharedService]()
			if service == nil {
				t.Error("Got nil service in concurrent resolution")
			}
		}()
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}

	close(errors)
	for err := range errors {
		if err != nil {
			t.Errorf("Unexpected error in concurrent resolution: %v", err)
		}
	}
}
