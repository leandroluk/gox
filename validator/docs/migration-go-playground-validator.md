# Migrating from go-playground/validator to v

In `go-playground/validator`, you define validation using **struct tags**. In `v`, you define it in **Go code**.
There are no "magic tags". Schemas are explicit. It is more verbose initially but much less mysterious and easier to debug.

## Key Differences
- **Missing vs Null**: In `v`, `missing` (not provided) and `null` (explicit null) are distinct states. By default, both pass unless `.Required()` is used.
- **Coercion is Opt-in**: `WithCoerce(true)` enables basic coercion; implicit type juggling is avoided.
- **Dedicated Schemas**: You choose the specific schema for the type (`Text()`, `Number()`, etc.), rather than a generic validator.
- **No `dive`**: For arrays/maps, you simply use `.Items(...)` or `.Values(...)`.

## Mental Model (Mapping Tags)
1. Find the **real type**: string → `Text()`, int/float → `Number[T]()`, bool → `Boolean()`, time → `Date()`, duration → `Duration()`, slices → `Array[E]()`, map → `Record[V]()`, struct → `Object(...)`.
2. Convert **presence** (`required`/`omitempty`) and **limits** (`min/max/len`).
3. Convert **specific validations** (email, uuid, ip, etc.) to `Text` schema methods.
4. For **cross-field** logic, use object-level rules or `Custom(rule)`.

---

## Table: Control Tags

| Tag (validator.v10) | Equivalent in `v`               | Notes                                                                    |
| ------------------- | ------------------------------- | ------------------------------------------------------------------------ |
| `required`          | `.Required()`                   | Checks presence (not just non-zero)                                      |
| `omitempty`         | `WithOmitZero(true)`            | Applies to reflected input. In JSON, "missing" is already handled.       |
| `omitnil`           | (handled via pointer + absence) | If field is `*T`, `nil` becomes `null`/missing depending on input source |
| `omitzero`          | `WithOmitZero(true)`            | Similar to `omitempty` behaviors                                         |
| `isdefault`         | `.IsDefault()`                  | "If zero-value, skip validations" (except `.Required()`)                 |
| `structonly`        | `object.StructOnly()`           | Only validation rules on the object itself                               |
| `nostructlevel`     | `object.NoStructLevel()`        | Only fields validation                                                   |
| `dive`              | `.Items(...)` / `.Values(...)`  | You define the item/value validator explicitly                           |
| `keys,endkeys`      | `.Record().Keys(...)`           | Validate map keys explicitly (rarely needed)                             |
| `                   | ` (OR)                          | `v.AnyOf(...)`                                                           | Combine independent schemas |

## Table: String Tags

| Tag              | Equivalent in `v` (Text) | Example                                     |
| ---------------- | ------------------------ | ------------------------------------------- |
| `min=3`          | `.Min(3)`                | `Text().Min(3)`                             |
| `max=50`         | `.Max(50)`               | `Text().Max(50)`                            |
| `len=8`          | `.Len(8)`                | `Text().Len(8)`                             |
| `eq=foo`         | `.Eq("foo")`             | `Text().Eq("foo")`                          |
| `ne=foo`         | `.Ne("foo")`             | `Text().Ne("foo")`                          |
| `startswith=ab`  | `.StartsWith("ab")`      | `Text().StartsWith("ab")`                   |
| `endswith=ab`    | `.EndsWith("ab")`        | `Text().EndsWith("ab")`                     |
| `contains=ab`    | `.Contains("ab")`        | `Text().Contains("ab")`                     |
| `excludes=ab`    | `.Excludes("ab")`        | `Text().Excludes("ab")`                     |
| `lowercase`      | `.Lowercase()`           | `Text().Lowercase()`                        |
| `uppercase`      | `.Uppercase()`           | `Text().Uppercase()`                        |
| `oneof=a b c`    | `.OneOf("a","b","c")`    | `Text().OneOf(...)`                         |
| `email`          | `.Email()`               | `Text().Email()`                            |
| `url`            | `.URL()`                 | `Text().URL()`                              |
| `http_url`       | `.HTTPURL()`             | `Text().HTTPURL()`                          |
| `uri`            | `.URI()`                 | `Text().URI()`                              |
| `urn_rfc2141`    | `.URNRFC2141()`          | `Text().URNRFC2141()`                       |
| `uuid` / `uuid4` | `.UUID()` / `.UUID4()`   | `Text().UUID4()`                            |
| `ip` / `ipv4`    | `.IP()` / `.IPv4()`      | `Text().IPv4()`                             |
| `cidr`           | `.CIDR()`                | `Text().CIDR()`                             |
| `mac`            | `.MAC()`                 | `Text().MAC()`                              |
| `hostname`       | `.Hostname()`            | `Text().Hostname()`                         |
| `numeric`        | `.Numeric()`             | Digits only                                 |
| `number`         | `.Number()`              | String containing valid number (no NaN/Inf) |
| `hexadecimal`    | `.Hexadecimal()`         | Accepts `0x` prefix                         |
| `base64`         | `.Base64()`              | Standard Base64                             |
| `credit_card`    | `.CreditCard()`          | Luhn check + Length                         |
| `semver`         | `.SemVer()`              | SemVer 2.0.0                                |
| `file` / `dir`   | `.File()` / `.Dir()`     | Exists on filesystem                        |

## Table: Number Tags

| Tag                 | Equivalent                                 | Notes                       |
| ------------------- | ------------------------------------------ | --------------------------- |
| `gte=10`            | `Number[T]().Min(10)`                      | `T` defines int/float logic |
| `lte=10`            | `Number[T]().Max(10)`                      |                             |
| `oneof=1 2 3`       | `.OneOf(1,2,3)`                            |                             |
| `min=3` (slice/map) | `Array[E]().Min(3)` / `Record[V]().Min(3)` | Depends on container schema |

---

## Examples

### 1. Before: Tags

```go
type User struct {
	Name string   `json:"name" validate:"required,min=3,max=50"`
	Age  int      `json:"age"  validate:"gte=0,lte=130"`
	Role string   `json:"role" validate:"oneof=admin user guest"`
	Tags []string `json:"tags" validate:"max=10,dive,min=2,max=20"`
}
```

### 1. After: Schema (Fluent API)

```go
type User struct {
	Name string   `json:"name"`
	Age  int      `json:"age"`
	Role string   `json:"role"`
	Tags []string `json:"tags"`
}

var UserSchema = v.Object(func(u *User, s *v.ObjectSchema[User]) {
	s.Field(&u.Name).Text().Required().Min(3).Max(50)
	s.Field(&u.Age).Number().Integer().Min(0).Max(130)
	s.Field(&u.Role).Text().OneOf("admin", "user", "guest")
	
	s.Field(&u.Tags).Array(v.Text().Min(2).Max(20)).Max(10)
})
```

### 2. `dive` + Maps

```go
// validate:"min=1,max=20,keys,alphanum,endkeys,required,dive,required,email"
var EmailsByKey = v.Record[string]().
	Min(1).
	Max(20).
	Keys(v.Text().Pattern("^[a-zA-Z0-9]+$")). // Validating keys
	Values(v.Text().Required().Email())       // Validating values
```

### 3. OR (`|`)

```go
s := v.AnyOf(
	v.Text().Email(),
	v.Text().UUID(),
)

out, err := s.Validate("john@example.com")
```

---

## Complex Scenarios

Some tags are highly specific or opinionated (e.g., `required_without_all`, `excluded_unless`).
In `v`, you handle these via:

- **Object rules**: Validate the whole struct and implement cross-field logic in Go.
- **Custom rules**: `s.Field(&u.X).Custom(func(ctx, val) ...)`
- **Conditonals**: If implemented, `RequiredIf` logic resides on the Object builder.

The rule is simple: if validation depends on **more than one field**, it typically belongs to the object schema, not the individual field schema.
