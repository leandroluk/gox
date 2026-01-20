package mut_test

import (
	"encoding/json"
	"testing"

	mut "github.com/leandroluk/gox/mut"
)

// User represents a complex struct for testing
type User struct {
	Name  mut.Mut[string] `json:"name"`
	Age   mut.Mut[int]    `json:"age,omitempty"`
	Email mut.Mut[string] `json:"-"` // Tag to use field name
	NoTag mut.Mut[bool]   // No tag at all
	Fixed string          `json:"fixed_data"` // Common field (not mut.Mut)
}

func TestMut_Core(t *testing.T) {
	t.Run("New without value", func(t *testing.T) {
		m := mut.New[string]()
		if m.Dirty() {
			t.Error("New() without params should not be dirty")
		}
	})

	t.Run("New with value", func(t *testing.T) {
		m := mut.New("test")
		if !m.Dirty() {
			t.Error("New() with param should be dirty")
		}
		if m.Get() != "test" {
			t.Errorf("Expected 'test', got %v", m.Get())
		}
	})

	t.Run("Set and GetAny", func(t *testing.T) {
		m := mut.New[int]()
		m.Set(100)
		if !m.Dirty() {
			t.Error("After Set(), should be dirty")
		}
		if m.GetAny() != 100 {
			t.Errorf("GetAny() should return 100, got %v", m.GetAny())
		}
	})
}

func TestMut_JSON(t *testing.T) {
	t.Run("Marshal", func(t *testing.T) {
		m := mut.New("hello")
		b, _ := json.Marshal(&m)
		if string(b) != `"hello"` {
			t.Errorf("Wrong marshal output: %s", string(b))
		}
	})

	t.Run("Unmarshal null", func(t *testing.T) {
		var m mut.Mut[string]
		err := json.Unmarshal([]byte(`null`), &m)
		if err != nil {
			t.Fatal(err)
		}
		if !m.Dirty() {
			t.Error("Unmarshal null should set dirty to true")
		}
	})
}

func TestToMap_Coverage(t *testing.T) {
	t.Run("Validations and Tags", func(t *testing.T) {
		u := User{}
		u.Name.Set("Leandro")
		u.Email.Set("test@test.com") // Tag "-"
		u.NoTag.Set(true)            // No tag

		// Testing pass-by-value (triggers reflect.New fallback in ToMap)
		res := mut.ToMap(u)

		if res["name"] != "Leandro" {
			t.Errorf("Custom tag 'name' failed, got %v", res["name"])
		}
		if res["Email"] != "test@test.com" {
			t.Errorf("Tag '-' should fallback to field name 'Email', got %v", res["Email"])
		}
		if res["NoTag"] != true {
			t.Errorf("Field without tag should use field name 'NoTag', got %v", res["NoTag"])
		}
		if len(res) != 3 {
			t.Errorf("Expected 3 fields, got %d", len(res))
		}
	})

	t.Run("Non-struct input", func(t *testing.T) {
		res := mut.ToMap(123) // Passing an int instead of a struct
		if len(res) != 0 {
			t.Error("ToMap with non-struct should return empty map")
		}
	})

	t.Run("Pointer to struct", func(t *testing.T) {
		u := &User{}
		u.Age.Set(30)
		res := mut.ToMap(u)
		if res["age"] != 30 {
			t.Errorf("Expected age 30, got %v", res["age"])
		}
	})
}
