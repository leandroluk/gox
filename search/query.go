package search

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
)

type FieldConstraint interface{ ~string }

type ProjectMode string

const (
	Include ProjectMode = "include"
	Exclude ProjectMode = "exclude"
)

type Project[K FieldConstraint] struct {
	Mode   ProjectMode `json:"mode"`
	Fields []K         `json:"fields"`
}

type SortOrder int8

const (
	ASC  SortOrder = 1
	DESC SortOrder = -1
)

type SortItem[K FieldConstraint] struct {
	Field K         `json:"field"`
	Order SortOrder `json:"order"`
}

// Query is the main structure for dynamic searches.
type Query[TType any, KeyList FieldConstraint] struct {
	Where   *TType              `json:"where,omitempty"`
	Project *Project[KeyList]   `json:"project,omitempty"`
	Sort    []SortItem[KeyList] `json:"sort,omitempty"`
	Limit   *int                `json:"limit,omitempty"`
	Offset  *int                `json:"offset,omitempty"`
}

// UnmarshalJSON implements custom logic to handle sort as an object in JSON
// but as a slice of SortItem internally to preserve order.
func (q *Query[TType, KeyList]) UnmarshalJSON(data []byte) error {
	// 1. Create an Alias to avoid infinite recursion during Unmarshal
	type Alias Query[TType, KeyList]

	// 2. In the auxiliary struct, define Sort as json.RawMessage.
	// This allows the default unmarshaler to accept the {} object without attempting
	// to convert it to a slice yet, avoiding a type mismatch error.
	aux := &struct {
		*Alias
		Sort json.RawMessage `json:"sort,omitempty"`
	}{Alias: (*Alias)(q)}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// 3. Now process the "sort" field manually using the token scanner
	// to preserve the order of the JSON object keys.
	dec := json.NewDecoder(bytes.NewReader(data))
	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if t == "sort" {
			t, err = dec.Token() // Expected: json.Delim '{'
			if err != nil || t != json.Delim('{') {
				break
			}

			// Clear the slice if the default unmarshal populated it (unlikely here)
			q.Sort = nil

			for dec.More() {
				key, err := dec.Token()
				if err != nil {
					return err
				}

				var order SortOrder
				if err := dec.Decode(&order); err != nil {
					return err
				}

				q.Sort = append(q.Sort, SortItem[KeyList]{
					Field: KeyList(key.(string)),
					Order: order,
				})
			}
			break
		}
	}

	return nil
}

// Validate checks if the projection fields exist in the TType struct tags.
func (q *Query[TType, KeyList]) Validate() error {
	if q.Project == nil || len(q.Project.Fields) == 0 {
		return nil
	}

	validFields := make(map[string]bool)
	typ := reflect.TypeFor[TType]()

	for i := 0; i < typ.NumField(); i++ {
		tag := typ.Field(i).Tag.Get("json")
		fieldName := strings.Split(tag, ",")[0]
		if fieldName != "" && fieldName != "-" {
			validFields[fieldName] = true
		}
	}

	for _, field := range q.Project.Fields {
		if !validFields[string(field)] {
			return fmt.Errorf("invalid projection field: %s", field)
		}
	}
	return nil
}

type Result[TType any] struct {
	Items  []TType `json:"items"`
	Total  int     `json:"total"`
	Offset int     `json:"offset"`
	Limit  int     `json:"limit"`
}
