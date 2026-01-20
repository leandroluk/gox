# oas

Type-safe OpenAPI 3.1 document builder for Go with 100% fluent API.

## Installation

```sh
go get github.com/leandroluk/gox/oas
```

## Quick Start

```go
package main

import (
  "encoding/json"
  "os"
  "github.com/leandroluk/gox/oas/types"
  "github.com/leandroluk/gox/oas/enums"
)

func main() {
  doc := types.New().
    OpenAPI("3.1.0").
    Info(func(i *types.Info) {
      i.Title("My API").Version("1.0.0")
    }).
    Path("/users", func(p *types.Path) {
      p.Get(func(o *types.Operation) {
        o.Summary("List users").
          Response("200", func(r *types.Response) {
            r.Description("Success").
              Json(func(m *types.MediaType) {
                m.Schema(func(s *types.Schema) {
                  s.Array().Items(func(item *types.Schema) {
                    item.Object().
                      Property("id", func(id *types.Schema) { id.String() }).
                      Property("name", func(name *types.Schema) { name.String() })
                  })
                })
              })
          })
      })
    })

  json.NewEncoder(os.Stdout).Encode(doc)
}
```

## ✨ Key Features

### 🎯 Fluent API with Void Callbacks

The entire API uses **void callbacks** for hierarchical construction:

```go
doc.Path("/products", func(p *types.Path) {
    p.Post(func(o *types.Operation) {
        o.Summary("Create product").
          RequestBody(func(rb *types.RequestBody) {
              rb.Json(func(m *types.MediaType) {
                  m.Schema(func(s *types.Schema) {
                      s.Object()
                  })
              })
          })
    })
})
```

### 🔐 Security Helpers (NestJS-style)

Declare schemes globally and use them in operations:

```go
// Document: register schemes
doc.WithBearerToken("bearer").
    WithApiKey("api_key", "header")

// Operation: use the schemes
operation.UseBearerToken("bearer", "read", "write").
          UseApiKey("api_key")
```

### 🏗️ Schema Helpers

Fluent methods for common types:

```go
schema.String()   // .Type(enums.SchemaString)
schema.Object()   // .Type(enums.SchemaObject)
schema.Array()    // .Type(enums.SchemaArray)
schema.Integer()  // .Type(enums.SchemaInteger)
schema.Number()   // .Type(enums.SchemaNumber)
schema.Boolean()  // .Type(enums.SchemaBoolean)
schema.Null()     // .Type(enums.SchemaNull)
```

### 📦 Content Type Helpers

Reduce verbosity for common content types:

```go
response.Json(func(m *types.MediaType) { ... })      // application/json
response.Xml(func(m *types.MediaType) { ... })       // text/xml
response.Html(func(m *types.MediaType) { ... })      // text/html
response.Plain(func(m *types.MediaType) { ... })     // text/plain

requestBody.Form(func(m *types.MediaType) { ... })       // x-www-form-urlencoded
requestBody.Multipart(func(m *types.MediaType) { ... })  // multipart/form-data
```

### ✅ Automatic Validation

All types validate required fields in `MarshalJSON`:

```go
doc := types.New()
// Without .Info() set
json.Marshal(doc) // ❌ Error: "Document.info is required"
```

### 🔄 Marshal + Unmarshal

Full support for serialization and parsing:

```go
// Marshal
data, _ := json.Marshal(doc)

// Unmarshal
var parsed types.Document
json.Unmarshal(data, &parsed)
```

## Complete Example: E-Commerce API

```go
package main

import (
  "encoding/json"
  "os"

  "github.com/leandroluk/gox/oas/enums"
  "github.com/leandroluk/gox/oas/types"
)

func main() {
  doc := types.New().
    OpenAPI("3.1.0").
    Info(func(i *types.Info) {
      i.Title("E-Commerce Platform API").
        Version("2.0.0").
        Description("Complete REST API for managing products, orders, and customers").
        Contact().Name("API Team").Email("api@ecommerce.com").URL("https://ecommerce.com")
      i.License().Name("MIT").URL("https://opensource.org/licenses/MIT")
    }).
    // Security Schemes
    WithBearerToken("bearer").
    WithApiKey("api_key", "header").
    // Servers
    Server("https://api.ecommerce.com/v2", func(s *types.Server) {
      s.Description("Production").
        Variable("env", func(v *types.ServerVariable) {
          v.Default("production").
            Enum("production", "staging").
            Description("Environment")
        })
    }).
    Server("http://localhost:3000", func(s *types.Server) {
      s.Description("Development")
    }).
    // Tags
    Tag("Products", func(t *types.Tag) {
      t.Description("Product catalog management").
        ExternalDoc("https://docs.ecommerce.com/products")
    }).
    Tag("Orders").      // Optional callback - just the name
    Tag("Customers").   // Optional callback - just the name
    // Components - Schemas
    Components(func(c *types.Components) {
      // Product schema
      c.Schema("Product", func(s *types.Schema) {
        s.Object().
          Required("id", "name", "price", "stock").
          Property("id", func(p *types.Schema) {
            p.String().
              Format("uuid").
              Example("550e8400-e29b-41d4-a716-446655440000").
              Description("Unique product identifier")
          }).
          Property("name", func(p *types.Schema) {
            p.String().
              MinLength(3).
              MaxLength(200).
              Example("Wireless Bluetooth Headphones").
              Description("Product display name")
          }).
          Property("description", func(p *types.Schema) {
            p.String().
              MaxLength(2000).
              Example("Premium noise-cancelling headphones with 30-hour battery life")
          }).
          Property("price", func(p *types.Schema) {
            p.Number().
              Minimum(0.01).
              Example(149.99).
              Description("Price in USD")
          }).
          Property("stock", func(p *types.Schema) {
            p.Integer().
              Minimum(0).
              Example(245).
              Description("Available inventory")
          }).
          Property("category", func(p *types.Schema) {
            p.String().Example("electronics")
          }).
          Property("tags", func(p *types.Schema) {
            p.Array().
              Items(func(item *types.Schema) { item.String() }).
              Example([]string{"audio", "wireless", "premium"})
          }).
          Property("active", func(p *types.Schema) {
            p.Boolean().Default(true)
          }).
          Property("images", func(p *types.Schema) {
            p.Array().
              MaxItems(10).
              Items(func(item *types.Schema) {
                item.Object().
                  Property("url", func(u *types.Schema) { u.String().Format("uri") }).
                  Property("alt", func(a *types.Schema) { a.String() })
              })
          }).
          Property("createdAt", func(p *types.Schema) {
            p.String().Format("date-time").ReadOnly(true)
          }).
          Property("updatedAt", func(p *types.Schema) {
            p.String().Format("date-time").ReadOnly(true)
          })
      })

      // ProductInput schema
      c.Schema("ProductInput", func(s *types.Schema) {
        s.Object().
          Required("name", "price").
          Property("name", func(p *types.Schema) {
            p.String().MinLength(3).MaxLength(200)
          }).
          Property("description", func(p *types.Schema) {
            p.String().MaxLength(2000)
          }).
          Property("price", func(p *types.Schema) {
            p.Number().Minimum(0.01)
          }).
          Property("stock", func(p *types.Schema) {
            p.Integer().Minimum(0).Default(0)
          }).
          Property("category", func(p *types.Schema) { p.String() }).
          Property("tags", func(p *types.Schema) {
            p.Array().Items(func(item *types.Schema) { item.String() })
          })
      })

      // Order schema
      c.Schema("Order", func(s *types.Schema) {
        s.Object().
          Required("id", "customerId", "items", "total", "status").
          Property("id", func(p *types.Schema) { p.String().Format("uuid") }).
          Property("customerId", func(p *types.Schema) { p.String().Format("uuid") }).
          Property("items", func(p *types.Schema) {
            p.Array().
              MinItems(1).
              Items(func(item *types.Schema) {
                item.Object().
                  Required("productId", "quantity", "price").
                  Property("productId", func(pi *types.Schema) { pi.String().Format("uuid") }).
                  Property("quantity", func(q *types.Schema) { q.Integer().Minimum(1) }).
                  Property("price", func(pr *types.Schema) { pr.Number() })
              })
          }).
          Property("total", func(p *types.Schema) { p.Number() }).
          Property("status", func(p *types.Schema) {
            p.String().Enum("pending", "paid", "shipped", "delivered", "cancelled")
          }).
          Property("createdAt", func(p *types.Schema) { p.String().Format("date-time") })
      })

      // Error schema
      c.Schema("Error", func(s *types.Schema) {
        s.Object().
          Required("error", "message").
          Property("error", func(p *types.Schema) {
            p.String().Example("VALIDATION_ERROR")
          }).
          Property("message", func(p *types.Schema) {
            p.String().Example("Invalid request parameters")
          }).
          Property("details", func(p *types.Schema) {
            p.Array().
              Items(func(item *types.Schema) {
                item.Object().
                  Property("field", func(f *types.Schema) { f.String() }).
                  Property("issue", func(i *types.Schema) { i.String() })
              })
          })
      })

      // Pagination schema
      c.Schema("PaginatedProducts", func(s *types.Schema) {
        s.Object().
          Property("data", func(p *types.Schema) {
            p.Array().Items(func(item *types.Schema) {
              item.Ref("#/components/schemas/Product")
            })
          }).
          Property("meta", func(p *types.Schema) {
            p.Object().
              Property("page", func(pg *types.Schema) { pg.Integer() }).
              Property("perPage", func(pp *types.Schema) { pp.Integer() }).
              Property("total", func(t *types.Schema) { t.Integer() }).
              Property("totalPages", func(tp *types.Schema) { tp.Integer() })
          })
      })
    })

  // Paths
  doc.
    // GET /products - List with pagination and filters
    Path("/products", func(p *types.Path) {
      p.Get(func(o *types.Operation) {
        o.Summary("List products").
          Description("Returns paginated list of products with optional filters").
          Tags("Products").
          Parameter(func(param *types.Parameter) {
            param.Name("page").
              In("query").
              Description("Page number").
              Schema(func(s *types.Schema) {
                s.Integer().Minimum(1).Default(1)
              })
          }).
          Parameter(func(param *types.Parameter) {
            param.Name("perPage").
              In("query").
              Description("Items per page").
              Schema(func(s *types.Schema) {
                s.Integer().Minimum(1).Maximum(100).Default(20)
              })
          }).
          Parameter(func(param *types.Parameter) {
            param.Name("category").
              In("query").
              Schema(func(s *types.Schema) { s.String() })
          }).
          Parameter(func(param *types.Parameter) {
            param.Name("search").
              In("query").
              Description("Search in name and description").
              Schema(func(s *types.Schema) { s.String() })
          }).
          Response("200", func(r *types.Response) {
            r.Description("Successful response").
              Json(func(m *types.MediaType) {
                m.Schema(func(s *types.Schema) {
                  s.Ref("#/components/schemas/PaginatedProducts")
                })
              })
          }).
          Response("400", func(r *types.Response) {
            r.Description("Invalid parameters").
              Json(func(m *types.MediaType) {
                m.Schema(func(s *types.Schema) {
                  s.Ref("#/components/schemas/Error")
                })
              })
          })
      })

      // POST /products - Create product (requires auth)
      p.Post(func(o *types.Operation) {
        o.Summary("Create product").
          Description("Creates a new product in the catalog").
          Tags("Products").
          UseBearerToken("bearer", "products:write").
          RequestBody(func(rb *types.RequestBody) {
            rb.Required(true).
              Description("Product data").
              Json(func(m *types.MediaType) {
                m.Schema(func(s *types.Schema) {
                  s.Ref("#/components/schemas/ProductInput")
                })
              }).
              Xml(func(m *types.MediaType) {
                m.Schema(func(s *types.Schema) {
                  s.Ref("#/components/schemas/ProductInput")
                })
              })
          }).
          Response("201", func(r *types.Response) {
            r.Description("Product created successfully").
              Json(func(m *types.MediaType) {
                m.Schema(func(s *types.Schema) {
                  s.Ref("#/components/schemas/Product")
                })
              })
          }).
          Response("400", func(r *types.Response) {
            r.Description("Validation error").
              Json(func(m *types.MediaType) {
                m.Schema(func(s *types.Schema) {
                  s.Ref("#/components/schemas/Error")
                })
              })
          }).
          Response("401", func(r *types.Response) {
            r.Description("Unauthorized")
          }).
          Response("409", func(r *types.Response) {
            r.Description("Product already exists")
          })
      })
    }).
    // GET /products/{id}
    Path("/products/{id}", func(p *types.Path) {
      p.Get(func(o *types.Operation) {
        o.Summary("Get product by ID").
          Tags("Products").
          Parameter(func(param *types.Parameter) {
            param.Name("id").
              In("path").
              Required(true).
              Description("Product ID").
              Schema(func(s *types.Schema) {
                s.String().Format("uuid")
              })
          }).
          Response("200", func(r *types.Response) {
            r.Description("Product found").
              Json(func(m *types.MediaType) {
                m.Schema(func(s *types.Schema) {
                  s.Ref("#/components/schemas/Product")
                }).
                Example(map[string]any{
                  "id":          "550e8400-e29b-41d4-a716-446655440000",
                  "name":        "Wireless Headphones",
                  "description": "Premium quality",
                  "price":       149.99,
                  "stock":       245,
                  "active":      true,
                })
              })
          }).
          Response("404", func(r *types.Response) {
            r.Description("Product not found").
              Json(func(m *types.MediaType) {
                m.Schema(func(s *types.Schema) {
                  s.Ref("#/components/schemas/Error")
                })
              })
          })
      })

      // PATCH /products/{id} - Update (requires auth)
      p.Patch(func(o *types.Operation) {
        o.Summary("Update product").
          Tags("Products").
          UseBearerToken("bearer", "products:write").
          Parameter(func(param *types.Parameter) {
            param.Name("id").
              In("path").
              Required(true).
              Schema(func(s *types.Schema) {
                s.String().Format("uuid")
              })
          }).
          RequestBody(func(rb *types.RequestBody) {
            rb.Required(true).
              Json(func(m *types.MediaType) {
                m.Schema(func(s *types.Schema) {
                  s.Ref("#/components/schemas/ProductInput")
                })
              })
          }).
          Response("200", func(r *types.Response) {
            r.Description("Updated successfully").
              Json(func(m *types.MediaType) {
                m.Schema(func(s *types.Schema) {
                  s.Ref("#/components/schemas/Product")
                })
              })
          }).
          Response("404", func(r *types.Response) {
            r.Description("Product not found")
          }).
          Response("401", func(r *types.Response) {
            r.Description("Unauthorized")
          })
      })

      // DELETE /products/{id} - Delete (requires auth)
      p.Delete(func(o *types.Operation) {
        o.Summary("Delete product").
          Tags("Products").
          UseBearerToken("bearer", "products:delete").
          Parameter(func(param *types.Parameter) {
            param.Name("id").
              In("path").
              Required(true).
              Schema(func(s *types.Schema) {
                s.String().Format("uuid")
              })
          }).
          Response("204", func(r *types.Response) {
            r.Description("Deleted successfully")
          }).
          Response("404", func(r *types.Response) {
            r.Description("Product not found")
          }).
          Response("401", func(r *types.Response) {
            r.Description("Unauthorized")
          })
      })
    }).
    // POST /orders - Create order (requires auth)
    Path("/orders", func(p *types.Path) {
      p.Post(func(o *types.Operation) {
        o.Summary("Create order").
          Description("Creates a new order for the authenticated customer").
          Tags("Orders").
          UseBearerToken("bearer", "orders:create").
          RequestBody(func(rb *types.RequestBody) {
            rb.Required(true).
              Json(func(m *types.MediaType) {
                m.Schema(func(s *types.Schema) {
                  s.Object().
                    Required("items").
                    Property("items", func(items *types.Schema) {
                      items.Array().
                        MinItems(1).
                        Items(func(item *types.Schema) {
                          item.Object().
                            Required("productId", "quantity").
                            Property("productId", func(id *types.Schema) {
                              id.String().Format("uuid")
                            }).
                            Property("quantity", func(q *types.Schema) {
                              q.Integer().Minimum(1).Maximum(100)
                            })
                        })
                    })
                })
              })
          }).
          Response("201", func(r *types.Response) {
            r.Description("Order created").
              Json(func(m *types.MediaType) {
                m.Schema(func(s *types.Schema) {
                  s.Ref("#/components/schemas/Order")
                })
              })
          }).
          Response("400", func(r *types.Response) {
            r.Description("Validation error")
          }).
          Response("401", func(r *types.Response) {
            r.Description("Unauthorized")
          })
      })
    })

  // Output
  encoder := json.NewEncoder(os.Stdout)
  encoder.SetIndent("", "  ")
  if err := encoder.Encode(doc); err != nil {
    panic(err)
  }
}
```

## API Reference

### Document

```go
types.New().
    OpenAPI(version).                              // Default: "3.1.0"
    Info(func(i *Info) {}).   
    Server(url, optionalBuild...).                 // Build callback is optional
    Path(path, func(p *Path) {}).
    Components(func(c *Components) {}).
    Tag(name, optionalBuild...).                   // Build callback is optional
    ExternalDocs(func(e *ExternalDocs) {}).
    ExternalDoc(url, optionalBuild...).            // Shorthand with URL required
    Security(name, scopes...)                      // Add global security requirement
```

#### Optional Parameters Pattern

Many methods support optional build callbacks using variadic parameters:

```go
// With callback for configuration
doc.Tag("Products", func(t *Tag) {
    t.Description("Product management")
})

// Without callback - just the required parameter
doc.Tag("Orders")
doc.Tag("Customers")

// Server with optional configuration
doc.Server("https://api.example.com")
doc.Server("http://localhost:3000", func(s *Server) {
    s.Description("Development")
})

// ExternalDoc with optional description
doc.ExternalDoc("https://docs.example.com")
doc.ExternalDoc("https://docs.example.com", func(e *ExternalDocs) {
    e.Description("Complete API documentation")
})
```

#### Security Helpers
```go
doc.WithBearerToken(name)                              // Register Bearer JWT scheme
doc.WithApiKey(name, in)                               // Register API Key (header/query/cookie)
doc.WithSecurityScheme(name, func(s *SecurityScheme) {})  // Custom security scheme
```

### Operation

```go
operation.
    Summary(text).
    Description(text).
    Tags(tags...).
    OperationId(id).
    Deprecated(bool).
    Parameter(func(p *Parameter) {}).
    RequestBody(func(rb *RequestBody) {}).
    Response(code, func(r *Response) {}).
    ExternalDocs(func(e *ExternalDocs) {}).
    ExternalDoc(url, optionalBuild...).          // Shorthand with URL required
    Server(url, optionalBuild...)                // Server override (optional callback)
```

#### Security Helpers
```go
operation.UseBearerToken(name, scopes...)           // Apply Bearer auth
operation.UseApiKey(name)                           // Apply API key
operation.UseSecurityScheme(name, scopes...)        // Custom security requirement
```

### Schema

#### Type Helpers
```go
schema.String()    // Sets type to "string"
schema.Object()    // Sets type to "object"
schema.Array()     // Sets type to "array"
schema.Integer()   // Sets type to "integer"
schema.Number()    // Sets type to "number"
schema.Boolean()   // Sets type to "boolean"
schema.Null()      // Sets type to "null"

// Nullable Types (OAS 3.1)
schema.String().Nullable()   // type: ["string", "null"]
schema.Integer().Nullable()  // type: ["integer", "null"]
```

#### String Validation
```go
schema.MinLength(n).
       MaxLength(n).
       Pattern(regex).
       Format("email" | "uuid" | "uri" | "date" | "date-time" | ...)
```

#### Number Validation
```go
schema.Minimum(n).
       Maximum(n).
       ExclusiveMinimum(bool).
       ExclusiveMaximum(bool).
       MultipleOf(n)
```

#### Array Validation
```go
schema.MinItems(n).
       MaxItems(n).
       UniqueItems(bool).
       Items(func(s *Schema) {})
```

#### Object Structure
```go
schema.RequiredProperties(fields...).
       Property(name, func(s *Schema) {}).
       Required(name, func(s *Schema) {}).  // Define property and mark as required
       Optional(name, func(s *Schema) {}).  // Alias for Property() for clarity
       AdditionalProperties(value).
       MinProperties(n).
       MaxProperties(n)
```

##### Required & Optional

Use `Required` and `Optional` to avoid repeating field names:

```go
// Traditional approach - field name repeated
schema.Object().
    Property("name", func(s *Schema) { s.String() }).
    Property("email", func(s *Schema) { s.String() }).
    RequiredProperties("name", "email")

// New approach - no repetition
schema.Object().
    Required("name", func(s *Schema) { s.String() }).
    Required("email", func(s *Schema) { s.String() }).
    Optional("age", func(s *Schema) { s.Integer() })
```


#### Composition
```go
schema.AllOf(func(s *Schema) {}).   // Must match all schemas
schema.OneOf(func(s *Schema) {}).   // Must match exactly one
schema.AnyOf(func(s *Schema) {}).   // Must match at least one
schema.Not(func(s *Schema) {})      // Must NOT match schema
```

#### Metadata
```go
schema.Title(text).
       Description(text).
       Example(value).
       Default(value).
       ReadOnly(bool).
       WriteOnly(bool).
       Deprecated(bool)
```

### Response & RequestBody

#### Content Type Helpers
```go
// Response
response.Json(func(m *MediaType) {})       // application/json
response.Xml(func(m *MediaType) {})        // text/xml
response.Html(func(m *MediaType) {})       // text/html
response.Plain(func(m *MediaType) {})      // text/plain

// RequestBody
requestBody.Json(func(m *MediaType) {})       // application/json
requestBody.Xml(func(m *MediaType) {})        // text/xml
requestBody.Form(func(m *MediaType) {})       // application/x-www-form-urlencoded
requestBody.Multipart(func(m *MediaType) {})  // multipart/form-data
```

#### Generic Content
```go
response.Content(enums.ContentJson, func(m *MediaType) {})
requestBody.Content(enums.ContentXml, func(m *MediaType) {})
```

### Parameter

```go
parameter.Name(name).
          In("query" | "header" | "path" | "cookie").
          Required(bool).
          Description(text).
          Schema(func(s *Schema) {}).
          Example(value).
          Deprecated(bool).
          AllowEmptyValue(bool).    // query/cookie only
          AllowReserved(bool)       // query only
```

### Components

```go
components.
    Schema(name, func(s *Schema) {}).
    Response(name, func(r *Response) {}).
    Parameter(name, func(p *Parameter) {}).
    Example(name, func(e *ExampleObject) {}).
    RequestBody(name, func(rb *RequestBody) {}).
    Header(name, func(h *Header) {}).
    SecurityScheme(name, func(ss *SecurityScheme) {}).
    Link(name, func(l *Link) {}).
    Callback(name, func(cb Callback) {}).
    Path(name, func(p *Path) {})
```

## Best Practices

### 1. Component Reuse
Define common schemas in components and reference them:

```go
doc.Components(func(c *Components) {
    c.Schema("Error", func(s *Schema) {
        s.Object().
          Required("error", "message").
          Property("error", func(p *Schema) { p.String() }).
          Property("message", func(p *Schema) { p.String() })
    })
})

// Reference with $ref
schema.Ref("#/components/schemas/Error")
```

### 2. Separate Input/Output Schemas
```go
c.Schema("ProductInput", ...)  // POST/PATCH (no id, timestamps)
c.Schema("Product", ...)       // Responses (with id, createdAt, updatedAt)
```

### 3. Add Validation Constraints
Always add reasonable validation to prevent abuse:

```go
schema.String().MinLength(1).MaxLength(255)  // Not unlimited
schema.Integer().Minimum(0)                  // Non-negative
schema.Array().MaxItems(100)                 // Prevent abuse
```

### 4. Use Descriptive Operation IDs
```go
operation.OperationId("createUser")    // ✅ Clear and specific
operation.OperationId("create")        // ❌ Ambiguous
```

### 5. Leverage Automatic Validation
```go
// Validation happens automatically during marshaling
data, err := json.Marshal(doc)
if err != nil {
    // Handle validation errors
    log.Fatal(err)
}
```

### 6. Security Best Practices
```go
// Register schemes once at document level
doc.WithBearerToken("bearer").
    WithApiKey("api_key", "header")

// Apply to specific operations
operation.UseBearerToken("bearer", "read", "write")

// Multiple schemes for an operation
operation.UseBearerToken("bearer").UseApiKey("api_key")
```

## Features

- ✅ **100% Fluent API** with void callbacks for intuitive hierarchy
- ✅ **Type-Safe** with strongly-typed enums
- ✅ **Automatic Validation** during JSON marshaling
- ✅ **Full Marshal + Unmarshal** support
- ✅ **Security Helpers** inspired by NestJS decorators
- ✅ **Schema Type Helpers** (.String(), .Object(), etc.)
- ✅ **Content Type Helpers** (.Json(), .Xml(), .Form(), etc.)
- ✅ **OpenAPI 3.1 Compliant** with default value handling
- ✅ **Zero External Dependencies**
- ✅ **Reuse-if-Exists** pattern for map builders
- ✅ **Comprehensive godoc** documentation

## Architecture

The package is organized into three main sub-packages:

- **`types`** - Core OpenAPI types with fluent builders
- **`enums`** - Type-safe enumerations (ContentType, SchemaType)
- **`oas`** (root) - Type aliases for backward compatibility

All types implement `json.Marshaler` and `json.Unmarshaler` for seamless JSON integration.

## License

MIT
