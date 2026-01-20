package search_test

import (
	"encoding/json"
	"testing"

	"github.com/leandroluk/gox/search"
)

type UserFilter struct {
	Name search.StringCondition      `json:"name"`
	Age  search.NumberCondition[int] `json:"age"`
}

type UserKeys string

type UserView struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestQuery_Unmarshal(t *testing.T) {
	jsonData := []byte(`{
		"where": { "name": { "like": "Leandro%" } },
		"sort": { "name": 1, "age": -1 },
		"limit": 10
	}`)

	var q search.Query[UserFilter, UserKeys]
	if err := json.Unmarshal(jsonData, &q); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if q.Where == nil || q.Where.Name.Like == nil || *q.Where.Name.Like != "Leandro%" {
		t.Error("Where clause not unmarshaled correctly")
	}

	if len(q.Sort) != 2 {
		t.Errorf("Expected 2 sort items, got %d", len(q.Sort))
	}

	if q.Sort[0].Field != "name" || q.Sort[0].Order != search.ASC {
		t.Error("First sort item incorrect")
	}
}

func TestQuery_ValidateAgainst(t *testing.T) {
	t.Run("Valid projection/sort against view", func(t *testing.T) {
		limit := 10
		q := search.Query[UserFilter, UserKeys]{
			Project: &search.Project[UserKeys]{
				Mode:   search.INCLUDE,
				Fields: []UserKeys{"name"},
			},
			Sort: []search.SortItem[UserKeys]{
				{Field: "age", Order: search.DESC},
			},
			Limit: &limit,
		}
		if err := q.ValidateAgainst(UserView{}); err != nil {
			t.Errorf("Should be valid: %v", err)
		}
	})

	t.Run("Invalid projection field", func(t *testing.T) {
		q := search.Query[UserFilter, UserKeys]{
			Project: &search.Project[UserKeys]{
				Mode:   search.INCLUDE,
				Fields: []UserKeys{"password"},
			},
		}
		if err := q.ValidateAgainst(UserView{}); err == nil {
			t.Error("Should have failed for invalid field 'password'")
		}
	})

	t.Run("Invalid sort field", func(t *testing.T) {
		q := search.Query[UserFilter, UserKeys]{
			Sort: []search.SortItem[UserKeys]{
				{Field: "password", Order: search.ASC},
			},
		}
		if err := q.ValidateAgainst(UserView{}); err == nil {
			t.Error("Should have failed for invalid sort field 'password'")
		}
	})

	t.Run("Invalid limit/offset", func(t *testing.T) {
		limit := -1
		q := search.Query[UserFilter, UserKeys]{Limit: &limit}
		if err := q.ValidateAgainst(UserView{}); err == nil {
			t.Error("Should have failed for negative limit")
		}
	})
}
