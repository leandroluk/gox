# Migrating from ozzo-validation to v

`ozzo-validation` is a popular library that validates structs by defining rules for fields.
`v` shares some similarities (explicit rules in Go) but differs in **how** schemas are constructed and reused.

## Key Differences

| Feature             | ozzo-validation                            | v                                                      |
| :------------------ | :----------------------------------------- | :----------------------------------------------------- |
| **Paradigm**        | Method chaining on fields                  | Fluent definition of Type Schema                       |
| **Separation**      | Validation logic mixed with struct (often) | Schema completely decoupled from Struct                |
| **Missing vs Null** | Often conflated                            | Explicit distinction (AST-based)                       |
| **Generics**        | `interface{}` based                        | Strongly types (`v.Object[T]`)                         |
| **Performance**     | Reflection on every call                   | Reflection once (at build time) for structure analysis |

## Mental Model Shift

In **Ozzo**, you usually implement `Validate()` on your struct:

```go
func (a Address) Validate() error {
    return validation.ValidateStruct(&a,
        validation.Field(&a.Street, validation.Required, validation.Length(5, 50)),
        validation.Field(&a.City, validation.Required),
        validation.Field(&a.State, validation.Required, validation.Match(regexp.MustCompile("^[A-Z]{2}$"))),
        validation.Field(&a.Zip, validation.Required, validation.Match(regexp.MustCompile("^[0-9]{5}$"))),
    )
}
```

In **v**, you define a **Schema** variable (usually globally or in a package) that knows how to validate that type:

```go
var AddressSchema = v.Object(func(a *Address, s *v.ObjectSchema[Address]) {
    s.Field(&a.Street).Text().Required().Min(5).Max(50)
    s.Field(&a.City).Text().Required()
    s.Field(&a.State).Text().Required().Pattern("^[A-Z]{2}$")
    s.Field(&a.Zip).Text().Required().Pattern("^[0-9]{5}$")
})
```

Then you call:
```go
address, err := AddressSchema.Validate(input)
```

## Migration Table

| Ozzo Rule                     | v Equivalent                           |
| :---------------------------- | :------------------------------------- |
| `validation.Required`         | `.Required()`                          |
| `validation.NilOrNotEmpty`    | (Default behavior if not required)     |
| `validation.Length(min, max)` | `.Min(min).Max(max)` / `.Len(n)`       |
| `validation.In(...)`          | `.OneOf(...)`                          |
| `validation.Match(re)`        | `.Pattern(str)` / `.PatternRegexp(re)` |
| `validation.Date(layout)`     | `v.Date()` (with `WithDateLayouts`)    |
| `is.Email`                    | `.Email()`                             |
| `is.URL`                      | `.URL()`                               |
| `is.UUID`                     | `.UUID()`                              |
| `is.Int`                      | `v.Number[int]()`                      |
| `is.Float`                    | `v.Number[float64]()`                  |

## Conditional Validation

Ozzo uses `.When(cond, rules...)`.
`v` handles this via:
1. **Object-level rules**: Implement checks inside the schema builder using standard Go `if`.
2. **Custom rules**: `.Custom(func(ctx, val) ...)` where you have full control.

## Nested Validation

**Ozzo:**
```go
validation.Field(&u.Address) // calls Address.Validate() automatically
```

**v:**
```go
s.Field(&u.Address).Object(AddressBuilder) // Explicitly defined or reuse a variable
```
Or reuse a registered schema:
```go
s.Field(&u.Address).Object(func(a *Address, s *v.ObjectSchema[Address]) {
    // define inline or look up
})
```
(Currently `v` prefers explicit schema definition over implicit `Validate()` method detection to maintain separation of concerns).
