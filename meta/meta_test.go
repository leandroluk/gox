package meta_test

import (
	"reflect"
	"testing"

	"github.com/leandroluk/gox/meta"
)

// Helper types for tests
type SubStruct struct {
	Details string `json:"details"`
}

type MainStruct struct {
	ID      string     `json:"id"`
	Nested  SubStruct  `json:"nested"`
	Pointer *SubStruct `json:"pointer"`
}

type MyError struct{}

func (e MyError) Error() string { return "Unauthorized access" }

type EnumType string

func (e EnumType) Values() []string { return []string{"A", "B", "C"} }

type ConstrainedStruct struct {
	Value    float64  `json:"val"`
	Text     string   `json:"txt"`
	Category EnumType `json:"cat"`
	Secret   string   `json:"-"`
}

type AnonBase struct {
	BaseField string `json:"base"`
}

type AnonStruct struct {
	AnonBase
	Value string `json:"val"`
}

func TestMeta_Resolution(t *testing.T) {
	m := &MainStruct{Pointer: &SubStruct{}}

	meta.Describe(m,
		meta.Description("Main Object"),
		meta.Example(MainStruct{ID: "123"}),
		meta.Field(&m.ID, meta.Description("The ID Field"), meta.Example("ABC")),
		meta.Field(&m.Nested.Details, meta.Description("Nested Detail")),
		meta.Field(&m.Pointer.Details, meta.Description("Pointer Detail")),
		// Trigger pointer element type branch in Field decorator
		meta.Field(&m.Pointer, meta.Description("Pointer object")),
	)

	t.Run("Object Metadata", func(t *testing.T) {
		data := meta.GetObjectMetadataAs[MainStruct]()
		if data == nil || data.Description != "Main Object" {
			t.Fatal("Metadata not found or incorrect")
		}
	})

	t.Run("Field Resolution Path", func(t *testing.T) {
		data := meta.GetObjectMetadataAs[MainStruct]()
		cases := []struct {
			path     string
			expected string
		}{
			{"ID", "The ID Field"},
			{"Nested.Details", "Nested Detail"},
			{"Pointer.Details", "Pointer Detail"},
		}
		for _, c := range cases {
			f, exists := data.Fields[c.path]
			if !exists || f.Description != c.expected {
				t.Errorf("Field %s mismatch", c.path)
			}
		}
	})
}

func TestMeta_Throws(t *testing.T) {
	meta.Describe(&MainStruct{}, meta.Throws[MyError]("Unauthorized access"))
	data := meta.GetObjectMetadataAs[MainStruct]()
	if len(data.Throws) == 0 || data.Throws[0].Description != "Unauthorized access" {
		t.Fatal("Throws metadata not registered correctly")
	}
}

func TestMeta_Constraints(t *testing.T) {
	s := &ConstrainedStruct{}
	meta.Describe(s,
		meta.Title("Constrained Object"),
		meta.Field(&s.Value, meta.Min(0), meta.Max(100), meta.Required()),
		meta.Field(&s.Text, meta.Pattern("^[a-z]+$"), meta.MinLength(1), meta.MaxLength(10), meta.WriteOnly()),
		meta.Field(&s.Category, meta.Description("Enum field")),
		meta.Field(&s.Secret, meta.ReadOnly(), meta.Deprecated()),
	)

	data := meta.GetObjectMetadataAs[ConstrainedStruct]()

	t.Run("Object Title", func(t *testing.T) {
		if data.Title != "Constrained Object" {
			t.Errorf("Title mismatch: %s", data.Title)
		}
	})

	t.Run("Required List", func(t *testing.T) {
		if len(data.Required) != 1 || data.Required[0] != "val" {
			t.Errorf("Required list mismatch: %v", data.Required)
		}
	})

	t.Run("Auto Enum", func(t *testing.T) {
		fCat := data.Fields["Category"]
		if len(fCat.Enum) != 3 {
			t.Fatal("Enum not auto-detected")
		}
	})

	t.Run("Visibility", func(t *testing.T) {
		fSec := data.Fields["Secret"]
		if !fSec.ReadOnly || !fSec.Deprecated {
			t.Error("Visibility mismatch")
		}
	})
}

func TestMeta_Registry_Gaps(t *testing.T) {
	type RegistryStruct struct{ Name string }
	s := &RegistryStruct{}
	meta.Describe(s, meta.Description("Registry Test"))

	t.Run("GetObjectMetadataAs - Non-pointer", func(t *testing.T) {
		data := meta.GetObjectMetadataAs[RegistryStruct]()
		if data == nil || data.Description != "Registry Test" {
			t.Error("Failed to get metadata using non-pointer type")
		}
	})

	t.Run("GetObjectMetadataOf", func(t *testing.T) {
		if meta.GetObjectMetadataOf(nil) != nil {
			t.Error("Nil check fail")
		}
		if meta.GetObjectMetadataOf(s) == nil {
			t.Error("Pointer check fail")
		}
		if meta.GetObjectMetadataOf(RegistryStruct{}) == nil {
			t.Error("Value check fail")
		}
		if meta.GetObjectMetadataOf(123) != nil {
			t.Error("Non-struct check fail")
		}
	})

	t.Run("GetObjectMetadataByType", func(t *testing.T) {
		if meta.GetObjectMetadataByType(nil) != nil {
			t.Error("Nil type fail")
		}
		if meta.GetObjectMetadataByType(reflect.TypeOf(s)) == nil {
			t.Error("Pointer type fail")
		}
	})
}

func TestMeta_Panics(t *testing.T) {
	assertPanic := func(t *testing.T, f func()) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Did not panic")
			}
		}()
		f()
	}
	t.Run("Describe nil", func(t *testing.T) { assertPanic(t, func() { meta.Describe(nil) }) })
	t.Run("Describe non-pointer", func(t *testing.T) { assertPanic(t, func() { meta.Describe(MainStruct{}) }) })
	t.Run("Describe non-struct pointer", func(t *testing.T) { v := 1; assertPanic(t, func() { meta.Describe(&v) }) })
	t.Run("Field invalid", func(t *testing.T) {
		assertPanic(t, func() { meta.Describe(&MainStruct{}, meta.Field("not a pointer")) })
	})
}

func TestMeta_AdditionalDecorators(t *testing.T) {
	type ExtraStruct struct {
		F1 string         `json:"f1"`
		M1 map[string]any `json:"m1"`
		I1 any            `json:"i1"`
	}
	s := &ExtraStruct{}
	meta.Describe(s,
		meta.Deprecated(),
		meta.ExtDocs("Doc", "url"),
		meta.Field(&s.F1,
			meta.Format("fmt"), meta.MultipleOf(2), meta.MinItems(1), meta.MaxItems(10),
			meta.ExtDocs("fdoc", "furl"), meta.Required(),
		),
		// Trigger alreadyPresent loop for Required
		meta.Field(&s.F1, meta.Required()),
		// Trigger Map and Interface branches in Field decorator
		meta.Field(&s.M1, meta.Description("Map")),
		meta.Field(&s.I1, meta.Description("Interface")),
	)
	data := meta.GetObjectMetadataAs[ExtraStruct]()
	if !data.Deprecated || data.ExternalDocs.URL != "url" {
		t.Error("Object metadata fail")
	}
	f1 := data.Fields["F1"]
	if f1.Format != "fmt" || *f1.MultipleOf != 2 || *f1.MinItems != 1 || *f1.MaxItems != 10 {
		t.Error("Field meta fail")
	}

	t.Run("Map and Interface Nullable", func(t *testing.T) {
		if !data.Fields["M1"].Nullable || !data.Fields["I1"].Nullable {
			t.Error("Map or Interface should be marked as nullable")
		}
	})

	t.Run("GetObjectMetadataAs with pointer", func(t *testing.T) {
		d := meta.GetObjectMetadataAs[*ExtraStruct]()
		if d == nil {
			t.Error("Failed with pointer type")
		}
	})
}

func TestMeta_Anonymous(t *testing.T) {
	s := &AnonStruct{}
	meta.Describe(s, meta.Field(&s.BaseField, meta.Description("Base")))
	data := meta.GetObjectMetadataAs[AnonStruct]()
	if data.Fields["BaseField"].Description != "Base" {
		t.Error("Anon resolution fail")
	}

	t.Run("ResolveFieldName - Non-struct target", func(t *testing.T) {
		val := 123
		name, _, _ := meta.ResolveFieldName(val, &val)
		if name != "" {
			t.Error("Should not resolve name for non-struct")
		}
	})
}

func TestMeta_DescribeMixedOptions(t *testing.T) {
	type Mixed struct{ F string }
	s := &Mixed{}
	// Test nil option
	meta.Describe(s, nil, meta.Description("D"))
	data := meta.GetObjectMetadataAs[Mixed]()
	if data.Description != "D" {
		t.Error("Mixed options fail")
	}
}
