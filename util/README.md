# Util

Package `util` provides atomic, generic utility functions to simplify common Go patterns, reducing boilerplate for error handling and type assertions.

## Functions

### Must

`Must` is a helper designed to wrap function calls that return `(T, error)` (or compatible types). It panics if the provided error is non-nil. If no error occurs, it asserts the value to type `T` and returns it.

This is particularly useful for global variable initialization or scenarios where an error should halt execution immediately.

**Signature:**
```go
func Must[T any](value any, err error) T
```

**Example:**
```go
// Assuming loadConfig returns (Config, error)
cfg := util.Must[Config](loadConfig())
```

### Ptr

`Ptr` performs a type assertion on a generic `any` value, returning it as type `T`.

**Signature:**
```go
func Ptr[T any](value any) T
```

**Example:**
```go
var data any = "example"
str := util.Ptr[string](data) // "example"
```
