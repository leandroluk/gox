# Dependency Injection (DI) Container

A lightweight, reflection-based Dependency Injection container for Go, supporting **Generics**, **Singleton/Transient** lifecycles, and **Interface Binding**.

## Features

- **Generic Support**: Type-safe resolution using Go Generics.
- **Lifecycles**: Support for Singletons (one instance) and Transients (new instance per resolution).
- **Dependency Graph**: Automatically resolves nested dependencies by analyzing factory function signatures.
- **Interface Binding**: Register concrete implementations as specific interfaces.
- **Concurrency Safe**: Thread-safe registry and instance caching using `sync.RWMutex`.

## Installation

```go
import "github.com/leandroluk/gox/di"
```