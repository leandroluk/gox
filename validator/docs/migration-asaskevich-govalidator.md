# Migrating from asaskevich/govalidator to v

`asaskevich/govalidator` is a classic Go library acting as both a **struct validator** (via tags) and a **string utility** belt (e.g. `IsEmail(str)`).
`v` upgrades this with **Type Safety**, strict **Separation of Concerns**, and a consistent **Fluent API**.

## 1. String Utilities

`govalidator` is often used as a direct string checker.
In `v`, you create a lightweight schema and call `.Validate()`.

| Feature    | govalidator                 | validator                                  |
| :--------- | :-------------------------- | :----------------------------------------- |
| **Email**  | `govalidator.IsEmail(str)`  | `validator.Text().Email().Validate(str)`   |
| **URL**    | `govalidator.IsURL(str)`    | `validator.Text().URL().Validate(str)`     |
| **UUID**   | `govalidator.IsUUID(str)`   | `validator.Text().UUID().Validate(str)`    |
| **Int**    | `govalidator.IsInt(str)`    | `validator.Text().Numeric().Validate(str)` |
| **Base64** | `govalidator.IsBase64(str)` | `validator.Text().Base64().Validate(str)`  |
| **JSON**   | `govalidator.IsJSON(str)`   | (No direct equivalent, unmarshal instead)  |

> **Note**: `v` returns `(value, error)`. If you just want a boolean, check `err == nil`.

## 2. Struct Validation (Tags vs Code)

`govalidator` relies heavily on reflection and magic strings in struct tags.

### Before: govalidator

```go
import "github.com/asaskevich/govalidator"

type User struct {
    Name  string `valid:"required,alpha"`
    Email string `valid:"required,email"`
    Age   int    `valid:"range(0|130)"`
}

func validate(u *User) (bool, error) {
    return govalidator.ValidateStruct(u)
}
```

### After: v

Structs are clean. Rules are code.

```go
import v "github.com/leandroluk/go/validator"

type User struct {
    Name  string
    Email string
    Age   int
}

var UserSchema = v.Object(func(u *User, s *v.ObjectSchema[User]) {
    s.Field(&u.Name).Text().Required().Pattern("^[a-zA-Z]+$") // "alpha"
    s.Field(&u.Email).Text().Required().Email()
    s.Field(&u.Age).Number().Min(0).Max(130)
})

func validate(u *User) (*User, error) {
    return UserSchema.Validate(u)
}
```

## 3. Why Switch?

### Type Safety
**govalidator**: `valid:"range(0|130)"` - If you type `rnge`, it might fail silently or panic at runtime.
**v**: `Min(0).Max(130)` - Checked at compile time.

### Null handling
**govalidator**: Often unclear behavior on zero values vs missing.
**v**: Explicit AST distinction between `missing`, `null`, and `check zero`.

### Custom Rules
**govalidator**: `govalidator.CustomTypeTagMap.Set("custom", ...)` (Global state!).
**v**: `.Custom(func(ctx, val) ...)` (Local, scoped, thread-safe).

## Migration Table (Tags)

| govalidator Tag       | validator Equivalent                 |
| :-------------------- | :----------------------------------- |
| `required`            | `.Required()`                        |
| `email`               | `.Email()`                           |
| `url`                 | `.URL()`                             |
| `alpha`               | `.Pattern("^[a-zA-Z]+$")`            |
| `alphanum`            | `.Pattern("^[a-zA-Z0-9]+$")`         |
| `numeric`             | `.Numeric()`                         |
| `hexadecimal`         | `.Hexadecimal()`                     |
| `uuid`                | `.UUID()`                            |
| `int`                 | `v.Number[int]()`                    |
| `float`               | `v.Number[float64]()`                |
| `range(min,max)`      | `.Min(min).Max(max)`                 |
| `length(min,max)`     | `.Min(min).Max(max)` (Text or Array) |
| `runelength(min,max)` | `.Min(min).Max(max)` (Text)          |
| `in(a,b, c)`          | `.OneOf("a", "b", "c")`              |
