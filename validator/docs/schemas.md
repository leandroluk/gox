# Schemas

Quick reference guide with short examples.

## Core Principles
- **Default**: `missing` and `null` pass (no issue generated) by default.
- **Required**: `.Required()` enforces presence.
- **Default Value**: `.Default(...)` applies before validation.
- **Pipeline**: Default → Check Missing/Null → Parse/Coerce/Type → Constraints → Custom Rules.
- **Coercion**: Opt-in via `WithCoerce(true)`.

---

## Text

Uses `v.Text()`.

### Meta
- `.Required()`
- `.IsDefault()` (skips validation if string is zero-value/empty, except checks `Required`)
- `.Default(v)` / `.DefaultFunc(fn)`
- `.Custom(rule)`

### Length & Equality
- `.Len(n)`
- `.Min(n)` / `.Max(n)`
- `.Eq(v)` / `.Ne(v)`
- `.EqIgnoreCase(v)` / `.NeIgnoreCase(v)`
- `.OneOf(v1, v2, ...)`

### Content (substring/prefix/suffix)
- `.Contains(v)` / `.Excludes(v)`
- `.StartsWith(v)` / `.NotStartsWith(v)`
- `.EndsWith(v)` / `.NotEndsWith(v)`
- `.Lowercase()` / `.Uppercase()` (validates case, does not change it unless coerce is used presumably? No, usually validation)

### Regex & Formats
- `.Pattern(regex)` / `.PatternRegexp(*regexp.Regexp)`
- `.Email()`
- `.URL()`
- `.HTTPURL()`
- `.URI()`
- `.URNRFC2141()`

### IDs & Network
- `.UUID()` / `.UUID3()` / `.UUID4()` / `.UUID5()`
- `.IP()` / `.IPv4()` / `.IPv6()`
- `.CIDR()`
- `.MAC()`
- `.Hostname()` / `.FQDN()`
- `.Port()`

### Numeric Strings & Colors
- `.Numeric()` (digits only)
- `.Number()` (float/exp string, disallows `NaN/Inf`)
- `.Hexadecimal()` (allows `0x`)
- `.HexColor()` (`#fff` / `#ffffff`)
- `.RGB()` / `.RGBA()` / `.HSL()` / `.HSLA()` (CSS-like)

### Encoding
- `.Base64()`
- `.Base64URL()` / `.Base64RawURL()`
- `.DataURI()`
- `.ASCII()` / `.PrintASCII()` / `.Multibyte()`

### Standards/Docs
- `.CreditCard()` / `.LuhnChecksum()`
- `.ISBN()` / `.ISBN10()` / `.ISBN13()`
- `.ISSN()`
- `.E164()`
- `.SemVer()`
- `.CVE()`

### Filesystem
- `.File()` / `.Dir()` (must exist)
- `.FilePath()` / `.DirPath()` (format check; `DirPath` requires trailing separator if not existing)
- `.Image()` (must exist + basic image decode check)

### Hash Digests
validates **hex** (fixed length) or **base64**.
- `.MD4()` / `.MD5()`
- `.SHA1()` / `.SHA224()` / `.SHA256()` / `.SHA384()` / `.SHA512()`
- `.SHA512_224()` / `.SHA512_256()`
- `.SHA3_224()` / `.SHA3_256()` / `.SHA3_384()` / `.SHA3_512()`
- `.RIPEMD160()`
- `.BLAKE2B_256()` ...
- `.BLAKE2S_256()`

### Example

```go
s := v.Text().Required().Email()
out, err := s.Validate("john@example.com")
```

---

## Number

Uses `v.Number[T]()`.

- `.Required()`
- `.Min(n)` / `.Max(n)`
- `.Integer()`
- `.Positive()` / `.Negative()`
- `.OneOf(v1, v2, ...)`
- `.Default(v)` / `.DefaultFunc(fn)`
- `.Custom(rule)`

```go
s := v.Number[int]().Min(0).Max(130)
```

---

## Boolean

Uses `v.Boolean()`.

- `.Required()`
- `.True()` / `.False()`
- `.Default(v)`
- `.Custom(rule)`

---

## Date

Uses `v.Date()`.

- `.Required()`
- `.Min(t)` / `.Max(t)`
- `.After(t)` / `.Before(t)`
- `.Default(v)`
- `.Custom(rule)`

Input parsing controlled by `WithTimeLocation(...)` and `WithDateLayouts(...)`.

---

## Duration

Uses `v.Duration()`.

- `.Required()`
- `.Min(d)` / `.Max(d)`
- `.Default(v)`
- `.Custom(rule)`

---

## Array

Uses `v.Array[E]()`.

- `.Required()`
- `.Min(n)` / `.Max(n)` (length)
- `.Default(v)`
- `.Items(schema)` (validates elements)
- `.Unique()` / `.UniqueByHash(fn)` / `.UniqueByEqual(fn)`
- `.Custom(rule)`

```go
// Array of strings, min length 1, max length 5
s := v.Array[string]().Min(1).Max(5).Items(v.Text().Email())
```

---

## Record (Map)

Uses `v.Record[V]()`.

- `.Required()`
- `.Min(n)` / `.Max(n)` (size)
- `.Default(v)`
- `.Keys(schema)`
- `.Values(schema)`
- `.Custom(rule)`

```go
// map[string]int
s := v.Record[int]().Values(v.Number[int]().Positive())
```

---

## Object (Struct)

Uses `v.Object(func...)`.

- `.Required()`
- `.Default(v)`
- `.Field(&u.X).Type()...` (fluent definition)
- `.StructOnly()` / `.NoStructLevel()`
- Custom object-level rules via builder or `Custom(rule)`.
