# Package Env

A flexible environment variable loader for Go 1.25+ that supports variable expansion, inline comments, and strong typing.

## Features

- **Variable Expansion**: Use `${VAR}` syntax to compose variables from others.
- **Strong Typing**: Automatically convert strings to `int`, `bool`, `time.Time`, `Duration`, and `json.RawMessage`.
- **Clean Parsing**: Supports spaces around `=`, `export` prefix, and `#` or `//` comments.
- **Default Values**: Provide fallbacks easily via generics.

---

## Usage

### 1. Create a `.env` file
```env
DB_HOST     = localhost
DB_USER     = admin
DB_URL      = "postgres://${DB_USER}@${DB_HOST}:5432/db"
DEBUG       = true
TIMEOUT     = 5s
```

### 2. Load and Access
```go
env.Load(".env")

url := env.Get[string]("DB_URL")
debug := env.Get[bool]("DEBUG", false)
timeout := env.Get[time.Duration]("TIMEOUT")
```

---

## Installation
```bash
go get github.com/leandroluk/go/env
```
