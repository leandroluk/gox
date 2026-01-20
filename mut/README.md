# Package mut

A utility package for Go that provides the generic `Mut[T]` type to handle **Partial Updates**, **State Tracking**, and **Patch** operations via JSON.

## Features

- **Generic `Mut[T]`**: Works with any data type using Go Generics.
- **Dirty Tracking**: Explicitly tracks if a value was assigned or modified.
- **JSON Native**: Seamlessly integrates with `encoding/json`.
- **SQL Ready**: Convert structs to maps for `UPDATE` operations using `ToMap`.
- **Minimalist API**: Focused on the `Get`, `Set`, and `Dirty` contract via the `Mutable` interface.

---

## Installation

```sh
go get github.com/leandroluk/gox/mut
```

## Usage

```go
type UserUpdate struct {
    Name  mut.Mut[string] `json:"name"`
    Age   mut.Mut[int]    `json:"age"`
}
```

```go
func main() {
    userUpdate := &UserUpdate{}
    userUpdate.Name.Set("Leandro")
    userUpdate.Age.Set(30)

    if userUpdate.Name.Dirty() {
        fmt.Println("UserUpdate.Name:", userUpdate.Name.Get())
    }

    if userUpdate.Age.Dirty() {
        fmt.Println("UserUpdate.Age:", userUpdate.Age.Get())
    }

    data, _ := json.Marshal(userUpdate)
    fmt.Println("JSON:", string(data))
}
```