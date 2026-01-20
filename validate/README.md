# Package v

**Type-safe** data validation in Go, Zod/Joi style (code-based schemas), with defaults, **opt-in** coercion, errors with **paths**, and support for validating **structs** and **JSON** (`[]byte` / `json.RawMessage`).

## Why v?

### 1. AST-Based Validation (Missing vs Null)
Most Go validators conflate `zero values` (0, "") with `missing values`. `validate` builds an Abstract Syntax Tree (AST) of your input first.
- **Missing**: The field was not present in the input (e.g. JSON key missing).
- **Null**: The field was present but explicit `null`.
- **Value**: The field has a value (even if it's 0 or empty string).

This allows you to implement efficient PATCH APIs where "missing" means "don't touch" and "null" means "unset".

### 2. DDD Friendly (Schema Decoupled from Structs)
Your domain models (structs) remain pure. No `validate:"required"` tags polluting your entities.
Validation logic lives in the **Infrastructure** or **Presentation** layer, defined explicitly in Go code.

### 3. Type Safety (Generics + Reflection)
`validate` uses Go generics (`v.Object[User]`) to ensure your schema matches your struct. If you rename a field in the struct but forget the schema, validation might fail or panic fast (depending on configuration), but type reference is safe. Methods like `.Field(&u.Name)` use pointer analysis to link schema rules to struct fields safely.

## Installation

```bash
go get github.com/leandroluk/gox/v
```

## Summary
- [Package v](#package-v)
  - [Why v?](#why-v)
    - [1. AST-Based Validation (Missing vs Null)](#1-ast-based-validation-missing-vs-null)
    - [2. DDD Friendly (Schema Decoupled from Structs)](#2-ddd-friendly-schema-decoupled-from-structs)
    - [3. Type Safety (Generics + Reflection)](#3-type-safety-generics--reflection)
  - [Installation](#installation)
  - [Summary](#summary)
  - [Quick Start](#quick-start)
    - [Basic Primitives](#basic-primitives)
    - [Struct Validation (Fluent API)](#struct-validation-fluent-api)
  - [Features](#features)
    - [Defaults \& null](#defaults--null)
    - [Coercion (Opt-in)](#coercion-opt-in)
    - [Fluent Builders](#fluent-builders)
    - [Transformation](#transformation)
  - [Error Handling](#error-handling)
  - [Documentation](#documentation)

## Quick Start

### Basic Primitives

```go
package main

import (
    "fmt"
    v "github.com/leandroluk/gox/validate"
)

func main() {
    nameSchema := v.Text().Required().Min(3).Max(50)
    value, err := nameSchema.Validate("John Doe")
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println(value) // "John Doe"
}
```

### Struct Validation (Fluent API)

```go
type User struct {
    Name    string
    Age     int
    Tags    []string
    Address Address
}

type Address struct {
    City string
}

func main() {
    // Define schema
    schema := v.Object(func(u *User, s *v.ObjectSchema[User]) {
        s.Field(&u.Name).Text().Required().Min(3)
        s.Field(&u.Age).Number().Integer().Min(0).Max(130)
        
        // Nested Array
        s.Field(&u.Tags).Array(v.Text().Min(2)).Max(10)

        // Nested Object
        s.Field(&u.Address).Object(func(a *Address, s *v.ObjectSchema[Address]) {
            s.Field(&a.City).Text().Required()
        })
    })

    // Validate
    jsonInput := []byte(`{
        "name": "Jane", 
        "age": 25, 
        "tags": ["dev", "go"], 
        "address": {"city": "New York"}
    }`)
    
    user, err := schema.Validate[User](jsonInput)
    if err != nil {
        // err is v.ValidationError
        fmt.Println(err)
    }
    fmt.Printf("%+v\n", user)
}
```

## Features

### Defaults & null
By default, standard values apply to both `missing` and `null`.
To disable defaults for `null` (allow null to pass as zero value or error): `v.WithDefaultOnNull(false)`.

### Coercion (Opt-in)
`validate` does not coerce types blindly. You must opt-in.
```go
v.Number[int]().Validate("123", v.WithCoerce(true)) // 123
```
Flags for specific behavior:
- `WithCoerceTrimSpace(true)`: " 123 " -> 123
- `WithCoerceNumberUnderscore(true)`: "1_000" -> 1000

### Fluent Builders
- **Text**: `Required`, `Min/MaxLen`, `Pattern`, `Email`, `UUID`...
- **Number**: `Min/Max`, `Integer`, `Positive`...
- **Boolean**: `True`, `False`...
- **Date**: `Min/Max`, `After/Before`...
- **Array**: `Min/Max` (items), `Unique`, `Items(Schema)`...
- **Object**: `Field`, `StructOnly`, `NoStructLevel`...
- **Record**: `Min/Max` (keys), `Key(Schema)`, `Value(Schema)`...

### Transformation
Use `.Transform` to parse or convert values safely during the validation pass.

```go
s.Field(&t.Algorithm).Transform(func(val any) (any, error) {
    str, ok := val.(string)
    if !ok {
        return nil, errors.New("expected string")
    }
    // Mapping string to internal Enum/Type
    if method, ok := myMap[str]; ok {
        return method, nil
    }
    return nil, errors.New("unknown algorithm")
})
```

## Error Handling

Errors are returned as `v.ValidationError`, which contains a list of issues.
Each issue has:
- **Path**: `user.address.city` or `tags[0]`
- **Message**: "required", "too short"
- **Code**: `text.min`, `object.required`

## Documentation

- Migration from `go-playground/validate`: [`docs/migration-go-playground.md`](docs/migration-go-playground.md)
- Migration from `ozzo-validation`: [`docs/migration-ozzo-validator.md`](docs/migration-ozzo-validator.md)
- Migration from `asaskevich/govalidator`: [`docs/migration-asaskevich-govalidator.md`](docs/migration-asaskevich-govalidator.md)
- Schema reference: [`docs/schemas.md`](docs/schemas.md)

Full documentation is available via GoDoc and strict typing in your IDE.
