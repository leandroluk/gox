# Package Meta

A type-safe metadata and "decorator" simulation for Go 1.25+. It allows you to attach documentation, examples, and error specifications to structs and their fields using pointers instead of error-prone strings.

## Key Features

- **Ref-Based Resolution**: Uses memory addresses to identify fields, ensuring your documentation never goes out of sync with your code.
- **Single Source of Truth**: Document once, use for Swagger, GraphQL, gRPC, or Validation logic.
- **Nested Support**: Automatically resolves fields in embedded or nested structs.
- **Strongly Typed Examples**: Use generics to ensure examples match field types.

## Usage

### 1. Define and Describe your Struct

```go
type User struct {
    ID   string
    Age  int
}

func init() {
    var user User
    meta.Describe(&user,
        meta.Description("Represents a system user"),
        meta.Field(&user.ID, 
            meta.Description("Unique identifier"),
            meta.Example("usr_123"),
        ),
        meta.Throws[ErrNotFound]("When user does not exist"),
    )
}
```

### 2. Retrieve Metadata

```go
m := meta.GetObjectMetadataAs[User]()
fmt.Println(m.Description) // "Represents a system user"
fmt.Println(m.Fields["ID"].Description) // "Unique identifier"
```

## Why this approach?

Unlike struct tags, which are limited to strings and become unreadable when too long, `meta` allows:
- Multi-line descriptions.
- Complex example objects (not just strings).
- Linking real Go Error types to documentation.

## Installation

```sh
go get github.com/leandroluk/gox/meta
```
