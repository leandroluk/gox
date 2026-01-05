package search

import "time"

// StringCondition defines filters for string fields.
type StringCondition struct {
	Equal    *string  `json:"eq,omitempty"`
	Like     *string  `json:"like,omitempty"`
	NotEqual *string  `json:"neq,omitempty"`
	NotLike  *string  `json:"nlike,omitempty"`
	In       []string `json:"in,omitempty"`
	NotIn    []string `json:"nin,omitempty"`
}

// NumberCondition defines filters for numeric fields using generics.
type NumberCondition[T int | int64 | float64 | float32] struct {
	Equal             *T   `json:"eq,omitempty"`
	Greater           *T   `json:"gt,omitempty"`
	GreaterOrEqual    *T   `json:"gte,omitempty"`
	Less              *T   `json:"lt,omitempty"`
	LessOrEqual       *T   `json:"lte,omitempty"`
	NotEqual          *T   `json:"neq,omitempty"`
	NotGreater        *T   `json:"ngt,omitempty"`
	NotGreaterOrEqual *T   `json:"ngte,omitempty"`
	NotLess           *T   `json:"nlt,omitempty"`
	NotLessOrEqual    *T   `json:"nlte,omitempty"`
	In                *[]T `json:"in,omitempty"`
	NotIn             *[]T `json:"nin,omitempty"`
}

// BooleanCondition defines filters for boolean fields.
type BooleanCondition struct {
	Equal    *bool   `json:"eq,omitempty"`
	NotEqual *bool   `json:"neq,omitempty"`
	In       *[]bool `json:"in,omitempty"`
	NotIn    *[]bool `json:"nin,omitempty"`
}

// DateCondition defines filters for time.Time fields.
type DateCondition struct {
	Equal             *time.Time   `json:"eq,omitempty"`
	Greater           *time.Time   `json:"gt,omitempty"`
	GreaterOrEqual    *time.Time   `json:"gte,omitempty"`
	Less              *time.Time   `json:"lt,omitempty"`
	LessOrEqual       *time.Time   `json:"lte,omitempty"`
	NotEqual          *time.Time   `json:"neq,omitempty"`
	NotGreater        *time.Time   `json:"ngt,omitempty"`
	NotGreaterOrEqual *time.Time   `json:"ngte,omitempty"`
	NotLess           *time.Time   `json:"nlt,omitempty"`
	NotLessOrEqual    *time.Time   `json:"nlte,omitempty"`
	In                *[]time.Time `json:"in,omitempty"`
	NotIn             *[]time.Time `json:"nin,omitempty"`
}
