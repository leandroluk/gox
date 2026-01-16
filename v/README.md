# Package v

**Type-safe** data validation in Go, Zod/Joi style (code-based schemas), with defaults, **opt-in** coercion, errors with **paths**, and support for validating **structs** and **JSON** (`[]byte` / `json.RawMessage`).

## What it does (no magic)
- **Code-based schemas**: compose rules using methods (`Text().Min(3)`, etc).
- **Single pass**: `Validate` handles default → parse/coerce → validation and returns the final type.
- **Presence**: differentiates between `missing` (not provided) and `null` (explicitly null).
- **Error paths**: `user.name`, `items[0]`, `meta["a-b"]`.
- **FailFast / MaxIssues** via options.
- **Ready-to-use common validations** (e.g., `email`, `uuid`, `ip`, `base64`, `semver`, `file/dir/image`, etc).

## Installation

```bash
go get github.com/leandroluk/go/v
```

## Quick Start

### Primitives

```go
package main

import (
	"fmt"

	"github.com/leandroluk/go/v"
)

func main() {
	nameSchema := v.Text().
		Required().
		Min(3).
		Max(50)

	value, err := nameSchema.Validate("Jo")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(value)
}
```

### Validate[T] (struct + registry)

`Validate[T]` looks up a compatible schema for `T` in the registry.

```go
package main

import (
	"fmt"

	"github.com/leandroluk/go/v"
)

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	schema := v.Object(func(u *User, s *v.ObjectSchema[User]) {
		s.Field(&u.Name, func(ctx *v.Context, v any) (any, bool) {
			return v.Text().Required().Min(3).ValidateAny(v, ctx.Options)
		})
		s.Field(&u.Age, func(ctx *v.Context, v any) (any, bool) {
			return v.NumberSchemaOf[int]().Min(0).Max(130).ValidateAny(v, ctx.Options)
		})
	})

	out, err := schema.Validate[User]([]byte(`{"name":"John","age":30}`))
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%+v\n", out)
}
```

### Defaults

By default, defaults apply to both `missing` and `null`. To disable for `null`, use `WithDefaultOnNull(false)`.

```go
ageSchema := v.NumberSchemaOf[int]().Default(18).Min(0).Max(130)

a, _ := ageSchema.Validate(nil)
b, _ := ageSchema.Validate(nil, v.WithDefaultOnNull(false))
```

### Coerce (opt-in) + flags

`WithCoerce(true)` enables base coercion.
More "aggressive" coercions are only enabled with specific flags.

Common examples (depend on the schema):
- `WithCoerceTrimSpace(true)` (e.g., `" 12 "`).
- `WithCoerceNumberUnderscore(true)` (e.g., `"1_000"`).
- `WithCoerceDurationSeconds(true)` / `WithCoerceDurationMilliseconds(true)` (e.g., `5` becomes `5s`/`5ms`).

```go
n, err := v.NumberSchemaOf[int]().Validate(
  " 1_000 ",
  v.WithCoerce(true),
  v.WithCoerceTrimSpace(true),
  v.WithCoerceNumberUnderscore(true),
)
```

### OmitZero (similar to omitempty)

When `WithOmitZero(true)` is active, "zero" values in structs/maps are omitted during reflected input.

```go
out, err := v.Validate[User](user, v.WithOmitZero(true))
```

## Available Schemas (summary)

- `text`: required, isDefault, len/min/max, equals, pattern, oneOf, email/url/uri/urn, uuid, ip, base64, semver/cve, filesystem (file/dir/image), hashes etc.
- `number`: required, min/max, default, oneOf, coerce(stringNumber + flags)
- `boolean`: required, default, coerce(string|0/1)
- `date`: required, min/max, default, parse(layouts+location), coerce(optional)
- `duration`: required, min/max, default, parse(durationString), number = nanos (AST), Go number = only with seconds/millis flags
- `array`: required, min/max, default, items, unique, coerce(singleton)
- `record`: required, min/max, default, keys, values, unique
- `object`: required, default, fields (resolve pointer + json tag), rules, StructOnly/NoStructLevel, cross-field conditions
- `combinator`: `AnyOf`, `OneOf`

Full list and examples: `docs/schemas.md`.

## Options (global)

- `WithFailFast(bool)`
- `WithMaxIssues(int)`
- `WithDefaultOnNull(bool)` (default: `true`)
- `WithCoerce(bool)`
- `WithOmitZero(bool)`
- `WithTimeLocation(*time.Location)`
- `WithDateLayouts(...string)`
- coercion flags (if available in your build): trim space, underscore, unix seconds/millis, etc

## Errors

When validation fails, it returns a `ValidationError` (in `internal/issues`) containing a list of issues:

- `issue.code` (e.g., `text.min`, `number.type`)
- `issue.message` (e.g., `too short`, `expected number`)
- `issue.path` (e.g., `user.name`, `items[0]`, `meta["a-b"]`)
- `issue.meta` (e.g., `expected`, `actual`, `min`, `max`, `value`, `error`)

## Docs

- Migration from `go-playground/validator` (tags → schema): `docs/migration-go-playground.md`
- Schema reference: `docs/schemas.md`
