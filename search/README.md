# Package Search

A powerful generic query builder for Go 1.25+ designed to handle dynamic filtering, sorting, and projection through JSON.

## Features

- **Type-Safe Filters**: Specialized conditions for `String`, `Number`, `Boolean`, and `Date`.
- **Custom Sort Unmarshaling**: Supports JSON object syntax for sorting (`{"field": 1}`) while maintaining order internally.
- **Projection Validation**: Validates if requested fields exist in the target struct's JSON tags.
- **Generics**: Works with any filter struct and field key constraint.

---

## Usage

### 1. Define Filter and Keys
```go
type UserFilter struct {
    Name search.StringCondition `json:"name"`
    Age  search.NumberCondition[int] `json:"age"`
}

type UserKeys string // Usually matches JSON tags
```

### 2. Parse Query from JSON
```go
jsonData := []byte(`{
    "where": { "age": { "gte": 18 } },
    "sort": { "name": 1 },
    "project": { "mode": "include", "fields": ["name"] },
    "limit": 20
}`)

var q search.Query[UserFilter, UserKeys]
json.Unmarshal(jsonData, &q)
```

### 3. Validate Projection
```go
if err := q.Validate(); err != nil {
    log.Fatal("User requested invalid fields")
}
```

---

## Core Components



- **Conditions**: Each type (String, Number, etc.) has its own set of operators (`eq`, `gt`, `like`, `in`, etc.).
- **Query Struct**: The container for `Where` (filters), `Sort`, `Project`, and pagination (`Limit`/`Offset`).
- **Result Struct**: A standard wrapper for paginated responses containing items and total count.

---

## Installation
```bash
go get github.com/leandroluk/go/search
```