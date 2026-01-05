package search

import (
	"encoding/json"
	"testing"
)

type UserFilter struct {
	Name StringCondition      `json:"name"`
	Age  NumberCondition[int] `json:"age"`
}

type UserKeys string

func TestQuery_Unmarshal(t *testing.T) {
	jsonData := []byte(`{
		"where": { "name": { "like": "Leandro%" } },
		"sort": { "name": 1, "age": -1 },
		"limit": 10
	}`)

	var q Query[UserFilter, UserKeys]
	if err := json.Unmarshal(jsonData, &q); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if q.Where.Name.Like == nil || *q.Where.Name.Like != "Leandro%" {
		t.Error("Where clause not unmarshaled correctly")
	}

	if len(q.Sort) != 2 {
		t.Errorf("Expected 2 sort items, got %d", len(q.Sort))
	}

	if q.Sort[0].Field != "name" || q.Sort[0].Order != ASC {
		t.Error("First sort item incorrect")
	}
}

func TestQuery_Validate(t *testing.T) {
	type Schema struct {
		FullName string `json:"full_name"`
	}

	t.Run("Valid projection", func(t *testing.T) {
		q := Query[Schema, string]{
			Project: &Project[string]{
				Fields: []string{"full_name"},
			},
		}
		if err := q.Validate(); err != nil {
			t.Errorf("Should be valid: %v", err)
		}
	})

	t.Run("Invalid projection", func(t *testing.T) {
		q := Query[Schema, string]{
			Project: &Project[string]{
				Fields: []string{"password"},
			},
		}
		if err := q.Validate(); err == nil {
			t.Error("Should have failed for invalid field 'password'")
		}
	})
}
