package util_test

import (
	"errors"
	"testing"

	"github.com/leandroluk/gox/util"
)

func TestID(t *testing.T) {
	id1 := util.ID()
	id2 := util.ID()
	if id1 == "" {
		t.Error("expected non-empty ID")
	}
	if id1 == id2 {
		t.Error("expected unique IDs")
	}
}

func TestPtr(t *testing.T) {
	v := 42
	p := util.Ptr(v)
	if p == nil {
		t.Fatal("expected non-nil pointer")
	}
	if *p != v {
		t.Errorf("expected %v, got %v", v, *p)
	}
}

func TestMust(t *testing.T) {
	t.Run("no error", func(t *testing.T) {
		v := util.Must(42, nil)
		if v != 42 {
			t.Errorf("expected 42, got %v", v)
		}
	})
	t.Run("with error", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic")
			}
		}()
		util.Must(0, errors.New("test error"))
	})
}

func TestTry(t *testing.T) {
	t.Run("no panic", func(t *testing.T) {
		err := util.Try(func() {})
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("panic with error", func(t *testing.T) {
		expectedErr := errors.New("panic error")
		err := util.Try(func() {
			panic(expectedErr)
		})
		if err != expectedErr {
			t.Errorf("expected %v, got %v", expectedErr, err)
		}
	})

	t.Run("panic with string", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected outer panic")
			}
		}()
		util.Try(func() {
			panic("string panic")
		})
	})
}

func TestCheck(t *testing.T) {
	t.Run("no error", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("unexpected panic: %v", r)
			}
		}()
		util.Check(nil)
	})

	t.Run("with error", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic")
			}
		}()
		util.Check(errors.New("test error"))
	})
}

func TestMapMerge(t *testing.T) {
	m1 := map[string]int{"a": 1, "b": 2}
	m2 := map[string]int{"b": 3, "c": 4}
	m3 := map[string]int{"c": 5, "d": 6}

	result := util.MapMerge(m1, m2, m3)

	expected := map[string]int{"a": 1, "b": 3, "c": 5, "d": 6}
	if len(result) != len(expected) {
		t.Fatalf("expected len %d, got %d", len(expected), len(result))
	}
	for k, v := range expected {
		if result[k] != v {
			t.Errorf("expected key %s to be %v, got %v", k, v, result[k])
		}
	}
}

func TestSetDefault(t *testing.T) {
	t.Run("zero value", func(t *testing.T) {
		v := util.SetDefault("", "default")
		if v != "default" {
			t.Errorf("expected 'default', got '%v'", v)
		}
	})
	t.Run("non-zero value", func(t *testing.T) {
		v := util.SetDefault("value", "default")
		if v != "value" {
			t.Errorf("expected 'value', got '%v'", v)
		}
	})
}

func TestSetNil(t *testing.T) {
	t.Run("nil interface", func(t *testing.T) {
		var p any
		v := util.SetNil(&p, "default")
		if v != "default" {
			t.Errorf("expected 'default', got '%v'", v)
		}
		if p != "default" {
			t.Errorf("expected pointer to be set to 'default'")
		}
	})
	t.Run("non-nil interface", func(t *testing.T) {
		var p any = "value"
		v := util.SetNil(&p, "default")
		if v != "value" {
			t.Errorf("expected 'value', got '%v'", v)
		}
		if p != "value" {
			t.Errorf("expected pointer to remain 'value'")
		}
	})
}

func TestStructFromMap(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	input := map[string]any{
		"name": "Alice",
		"age":  30,
	}

	result, err := util.StructFromMap[Person](input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Name != "Alice" {
		t.Errorf("expected 'Alice', got '%s'", result.Name)
	}
	if result.Age != 30 {
		t.Errorf("expected 30, got %d", result.Age)
	}
}
