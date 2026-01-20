# Package Search

A generic query builder for Go 1.25+ designed to handle dynamic filtering, sorting, and projection through JSON.

## Features

- **Type-Safe Filters**: Specialized conditions for `String`, `Number`, `Boolean`, and `Date`.
- **Custom Sort Unmarshaling**: Supports JSON object syntax for sorting (`{"field": 1}`) while maintaining order internally.
- **Projection Validation**: Validates if requested projection/sort fields exist in the target struct's JSON tags.
- **Generics**: Works with any filter struct and field key constraint.

## Usage

### 1. Define Filter and Keys

```go
type UserFilter struct {
    Name search.StringCondition      `json:"name"`
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

### 3. Validate Projection and Sort

If your filter shape matches the returned item shape, you can call:

```go
if err := q.Validate(); err != nil {
    log.Fatal(err)
}
```

If you want to validate against a different struct (recommended for read models/views), use:

```go
type UserView struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

if err := q.ValidateAgainst[UserView](); err != nil {
    log.Fatal(err)
}
```

## Core Components

- **Conditions**: Each type (String, Number, etc.) has its own set of operators (`eq`, `gt`, `like`, `in`, etc.).
- **Query Struct**: The container for `Where` (filters), `Sort`, `Project`, and pagination (`Limit`/`Offset`).
- **Result Struct**: A standard wrapper for paginated responses containing items and total count.

## Installation

```sh
go get github.com/leandroluk/gox/search
```
