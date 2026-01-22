package meta

import (
	"testing"
)

type SubStruct struct {
	Details string
}

type MainStruct struct {
	ID      string
	Nested  SubStruct
	Pointer *SubStruct
}

type MyError struct{}

func (e MyError) Error() string {
	return "Unauthorized access"
}

func TestMeta_Resolution(t *testing.T) {
	// Instance used for address resolution
	// Note: We initialize pointers to ensure they are addressable in memory
	m := &MainStruct{Pointer: &SubStruct{}}

	Describe(m,
		Description("Main Object"),
		Example(MainStruct{ID: "123"}),
		// Direct field
		Field(&m.ID, Description("The ID Field"), Example("ABC")),
		// Nested struct field (Value) - This used to cause address collision with 'Nested'
		Field(&m.Nested.Details, Description("Nested Detail")),
		// Nested struct field (Pointer)
		Field(&m.Pointer.Details, Description("Pointer Detail")),
	)

	t.Run("Object Metadata", func(t *testing.T) {
		data := GetObjectMetadataAs[MainStruct]()
		if data == nil {
			t.Fatal("Metadata not found")
		}
		if data.Description != "Main Object" {
			t.Errorf("Expected 'Main Object', got %s", data.Description)
		}
	})

	t.Run("Field Resolution Path", func(t *testing.T) {
		data := GetObjectMetadataAs[MainStruct]()

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
			if !exists {
				t.Errorf("Field %s was not resolved by memory address", c.path)
				continue
			}
			if f.Description != c.expected {
				t.Errorf("Field %s: expected description %s, got %s", c.path, c.expected, f.Description)
			}
		}
	})
}

func TestMeta_Throws(t *testing.T) {

	Describe(&MainStruct{},
		Throws[MyError]("Unauthorized access"),
	)

	data := GetObjectMetadataAs[MainStruct]()
	if len(data.Throws) == 0 {
		t.Fatal("Throws metadata not registered")
	}

	if data.Throws[0].Description != "Unauthorized access" {
		t.Errorf("Incorrect Throws description")
	}
}
