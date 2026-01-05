# Leandro Luk - Go Core Libraries

This repository is a collection of high-performance, decoupled, and type-safe Go libraries designed for modern application development. 

All packages are compatible with **Go 1.25+** and leverage advanced features like Generics and optimized Reflection.

## Modules Overview

| Module                 | Purpose                                                                            | Status |
| :--------------------- | :--------------------------------------------------------------------------------- | :----: |
| [**di**](./di)         | Lightweight Dependency Injection container with Singleton/Transient support.       |   ✅    |
| [**cqrs**](./cqrs)     | Mediator for Commands and Queries with automatic type coercion.                    |   ✅    |
| [**set**](./set)       | Tracks partial JSON updates to distinguish between missing fields and zero-values. |   ✅    |
| [**search**](./search) | Generic query builder for complex filtering, sorting, and pagination.              |   ✅    |

## Project Structure

The project uses **Go Workspaces** to manage multiple modules. 

```text
.
├── cqrs/            # CQRS Mediator Module
├── di/              # Dependency Injection Module
└── search/          # Search & Query Module
├── set/             # Partial Update Module
├── go.work          # Workspace configuration
```

## Getting Started

To initialize the workspace and run tests across all modules:

```bash
go work init ./cqrs ./di ./search ./set
go test ./...
```

## Philosophy

- **Zero Dependencies**: Core modules aim for zero external dependencies.
- **Type Safety**: Use of Generics to avoid `interface{}` and runtime casting errors.
- **Convention over Configuration**: Smart defaults (like JSON tag reflection) to reduce boilerplate.
