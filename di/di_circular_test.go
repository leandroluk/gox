package di_test

import (
	"strings"
	"testing"

	"github.com/leandroluk/gox/di"
)

type CircularServiceA struct {
	B *CircularServiceB
}

type CircularServiceB struct {
	C *CircularServiceC
}

type CircularServiceC struct {
	A *CircularServiceA
}

type CircularDirectA struct {
	B *CircularDirectB
}

type CircularDirectB struct {
	A *CircularDirectA
}

type CircularSelfRef struct {
	Self *CircularSelfRef
}

func TestDI_CircularDependency_Direct(t *testing.T) {
	di.Reset()

	di.RegisterFrom[*CircularDirectA](func() (*CircularDirectA, error) {
		b := di.Resolve[*CircularDirectB]()
		return &CircularDirectA{B: b}, nil
	})
	di.RegisterFrom[*CircularDirectB](func() (*CircularDirectB, error) {
		a := di.Resolve[*CircularDirectA]()
		return &CircularDirectB{A: a}, nil
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

		if !strings.Contains(errMsg, "dependency chain") {
			t.Errorf("Error should show dependency chain, got: %s", errMsg)
		}

		if !strings.Contains(errMsg, "hint:") {
			t.Errorf("Error should include a hint, got: %s", errMsg)
		}
	}()

	di.Resolve[*CircularDirectA]()
}

func TestDI_CircularDependency_Indirect(t *testing.T) {
	di.Reset()

	di.RegisterFrom[*CircularServiceA](func() (*CircularServiceA, error) {
		b := di.Resolve[*CircularServiceB]()
		return &CircularServiceA{B: b}, nil
	})
	di.RegisterFrom[*CircularServiceB](func() (*CircularServiceB, error) {
		c := di.Resolve[*CircularServiceC]()
		return &CircularServiceB{C: c}, nil
	})
	di.RegisterFrom[*CircularServiceC](func() (*CircularServiceC, error) {
		a := di.Resolve[*CircularServiceA]()
		return &CircularServiceC{A: a}, nil
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

		hasServiceType := strings.Contains(errMsg, "CircularServiceA") ||
			strings.Contains(errMsg, "CircularServiceB") ||
			strings.Contains(errMsg, "CircularServiceC")

		if !hasServiceType {
			t.Errorf("Error should show types involved in cycle, got: %s", errMsg)
		}
	}()

	di.Resolve[*CircularServiceA]()
}

func TestDI_CircularDependency_WithSingleton(t *testing.T) {
	di.Reset()

	di.SingletonFrom[*CircularDirectA](func() (*CircularDirectA, error) {
		b := di.Resolve[*CircularDirectB]()
		return &CircularDirectA{B: b}, nil
	})
	di.SingletonFrom[*CircularDirectB](func() (*CircularDirectB, error) {
		a := di.Resolve[*CircularDirectA]()
		return &CircularDirectB{A: a}, nil
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

	di.RegisterFrom[*CircularSelfRef](func() (*CircularSelfRef, error) {
		self := di.Resolve[*CircularSelfRef]()
		return &CircularSelfRef{Self: self}, nil
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

	di.RegisterFrom[*CircularConfig](func() (*CircularConfig, error) {
		return &CircularConfig{Value: "test"}, nil
	})
	di.RegisterFrom[*CircularDatabase](func() (*CircularDatabase, error) {
		cfg := di.Resolve[*CircularConfig]()
		return &CircularDatabase{Config: cfg}, nil
	})
	di.RegisterFrom[*CircularService](func() (*CircularService, error) {
		db := di.Resolve[*CircularDatabase]()
		return &CircularService{DB: db}, nil
	})

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
	di.RegisterFrom[*CircularSharedService](func() (*CircularSharedService, error) {
		counter++
		return &CircularSharedService{ID: counter}, nil
	})

	const goroutines = 10
	done := make(chan bool, goroutines)
	errors := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					if errMsg, ok := r.(string); ok && strings.Contains(errMsg, "circular") {
						errors <- nil
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
