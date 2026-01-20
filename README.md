<p>
  <img alt="Build Status" src="https://github.com/leandroluk/gox/actions/workflows/ci.yml/badge.svg?branch=main">
  <img alt="Coverage" src=".public/coverage.svg">
  <img alt="Release" src="https://img.shields.io/github/release/leandroluk/go.svg?style=flat">
</p>

# Go Core Libraries

A collection of high-performance, decoupled, and type-safe Go libraries for modern application development.

## Contents
- [Go Core Libraries](#go-core-libraries)
  - [Contents](#contents)
  - [Getting started](#getting-started)
    - [Run tests across all modules](#run-tests-across-all-modules)
  - [Modules](#modules)
  - [Philosophy](#philosophy)
  - [Contributing](#contributing)
  - [License](#license)

## Getting started

This repository uses **Go Workspaces** (`go.work`) and supports **Go 1.25+**.

### Run tests across all modules

```sh
make test
```

## Modules

| Module                     | Purpose                                                                      |                         Coverage                         |
| :------------------------- | :--------------------------------------------------------------------------- | :------------------------------------------------------: |
| [**cqrs**](./cqrs)         | Mediator for Commands and Queries with automatic type coercion.              |     [![coverage](.public/cqrs-coverage.svg)](./cqrs)     |
| [**di**](./di)             | Lightweight Dependency Injection container with Singleton/Transient support. |       [![coverage](.public/di-coverage.svg)](./di)       |
| [**env**](./env)           | Environment variables parser with automatic type coercion.                   |      [![coverage](.public/env-coverage.svg)](./env)      |
| [**meta**](./meta)         | Metadata builder for complex filtering, sorting, and pagination.             |     [![coverage](.public/meta-coverage.svg)](./meta)     |
| [**mut**](./mut)           | Tracks partial JSON updates to distinguish missing fields from zero-values.  |      [![coverage](.public/mut-coverage.svg)](./mut)      |
| [**oas**](./oas)           | OpenAPI (Swagger) builder for complex filtering, sorting, and pagination.    |      [![coverage](.public/oas-coverage.svg)](./oas)      |
| [**search**](./search)     | Generic query builder for complex filtering, sorting, and pagination.        |   [![coverage](.public/search-coverage.svg)](./search)   |
| [**validate**](./validate) | Generic validator for complex filtering, sorting, and pagination.            | [![coverage](.public/validate-coverage.svg)](./validate) |

## Philosophy

- **Zero Dependencies**: core modules aim for zero external dependencies.
- **Type Safety**: heavy use of Generics to avoid `interface{}` and runtime casting errors.
- **Convention over Configuration**: smart defaults (like JSON tag reflection) to reduce boilerplate.

## Contributing

PRs are welcome. Keep changes scoped, tested, and consistent with the module style.

## License

MIT License — see [LICENSE](LICENSE).
