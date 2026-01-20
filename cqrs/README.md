# Package CQRS

A robust Command Query Responsibility Segregation (CQRS) mediator for Go 1.25+.

## Key Features

- **Decoupled Handlers**: Leverages the `di` package for lifecycle management.
- **Auto-Coercion**: Automatically handles pointer/value mismatches between dispatchers and handlers.
- **Type Safe**: Returns precise types using Go Generics.
- **High Performance**: RWMutex protected registry with optimized reflection lookups.

## Usage

### 1. Define your Query and Result

```go
type GetUserQuery struct { ID string }
type User struct { Name string }
```

### 2. Create the Handler

The handler must implement the `IQueryHandler` interface.

```go
type UserHandler struct {}

var _ cqrs.IQueryHandler[GetUserQuery, *User] = (*UserHandler)(nil)

func (h *UserHandler) Handle(ctx context.Context, q GetUserQuery) (*User, error) {
    return &User{Name: "Leandro Luk"}, nil
}
```

### 3. Register with DI Factory

Register the handler specifying the Message, the Result, and the Handler Type.

```go
cqrs.RegisterQueryHandler[GetUserQuery, *User, *UserHandler](func() *UserHandler {
    return &UserHandler{}
})
```

### 4. Dispatch

You can send the query as a value or a pointer; the engine will coerce it automatically.

```go
user, err := cqrs.ExecuteQuery[*User](ctx, GetUserQuery{ID: "123"})
```

## Technical Design

- **Normalization**: The registry normalizes types to ensure that `T` and `*T` resolve to the same handler.
- **Coercion Engine**: Before execution, the engine checks if the provided message matches the handler's input, performing pointer indirection or address-of operations if necessary.
- **DI Integration**: Handlers are retrieved from the `di` container, allowing them to have their own dependencies (repositories, clients, etc.) injected at construction time.


## Installation

```sh
go get github.com/leandroluk/gox/cqrs
```
