# Builder Generator

A Go code generator that creates builder pattern implementations for structs.

## Installation

```bash
go install github.com/nanostack-dev/generators/cmd/builder/generator@latest
```

## Usage

Add the `@builder` annotation to your struct and use `go:generate` to run the generator:

```go
//go:generate go run github.com/nanostack-dev/generators/cmd/builder/generator

// @builder
type Document struct {
    ID          string
    Title       string
    Description *string
    Content     string
}
```

## Annotations

You can customize the builder generation using annotations in the struct's documentation:

```go
// @builder
// @builder:prefix Set       // Use "Set" instead of "With" for method prefixes
// @builder:package models   // Override package name
// @builder:output {name}.generated.go  // Custom output file pattern
// @builder:immutable       // Generate immutable builder (Copy-on-write)
// @builder:nochain        // Methods return error instead of builder (default: false)
// @builder:constructor NewCustomBuilder  // Custom constructor name
// @builder:validate      // Generate validation methods
// @builder:skip         // Skip builder generation for this struct
// @builder:map Get:Build    // Map method names (e.g., Get() calls Build())
// @builder:custom UpdateCode  // Skip generation for WithUpdateCode() to implement it manually
```

### Custom Method Implementation

You can prevent the generator from creating specific builder methods using `@builder:custom`. This allows you to implement these methods manually with custom logic:

```go
// @builder
// @builder:custom UpdateCode  // Skip WithUpdateCode generation
type Document struct {
    Title      string
    Content    string
    UpdateCode *string
}

// Implement your custom method
func (b *DocumentBuilder) WithUpdateCode(code *string) *DocumentBuilder {
    // Custom implementation, e.g., hash the code
    hashedCode := hash(code)
    b.instance.UpdateCode = &hashedCode
    return b
}
```

### Method Mapping

You can create method aliases using the `@builder:map` annotation:

```go
// @builder
// @builder:map Get:Build        // Creates Get() that calls Build()
// @builder:map GetPtr:BuildAsPtr  // Creates GetPtr() that calls BuildAsPtr()
type Document struct {
    Title    string
    Content  string
}

// Usage:
doc := NewDocumentBuilder().
    WithTitle("Hello").
    WithContent("World").
    Get()  // Same as Build()

docPtr := NewDocumentBuilder().
    WithTitle("Hello").
    WithContent("World").
    GetPtr()  // Same as BuildAsPtr()
```

### Build Methods

Each builder includes two methods for obtaining the built instance:

```go
// Returns the struct by value
func (b *DocumentBuilder) Build() Document

// Returns the struct by pointer
func (b *DocumentBuilder) BuildAsPtr() *Document
```

### Method Chaining

By default, builder methods return the builder instance for method chaining:

```go
doc := NewDocumentBuilder().
    WithTitle("My Doc").
    WithContent("Some content").
    Build()
```

To disable method chaining and use error returns instead, use the `@builder:nochain` annotation:

```go
// @builder
// @builder:nochain
type Document struct {
    // ...
}

// Generated methods will return error:
if err := builder.WithTitle("My Doc"); err != nil {
    return err
}
```

### Examples

1. Basic usage with default settings:
```go
// @builder
type Document struct {
    Title    string
    Content  string
}

// Usage:
doc := NewDocumentBuilder().
    WithTitle("Hello").
    WithContent("World").
    Build()
```

2. Custom method names and implementations:
```go
// @builder
// @builder:prefix Set
// @builder:map Create:Build
// @builder:custom Password  // Skip SetPassword generation
type User struct {
    Username string
    Password string
}

// Custom password handling implementation
func (b *UserBuilder) SetPassword(password string) *UserBuilder {
    hashedPwd := hashPassword(password)
    b.instance.Password = hashedPwd
    return b
}

// Usage:
user := NewUserBuilder().
    SetUsername("john").
    SetPassword("secret").  // Uses custom implementation
    Create()               // Same as Build()
```

## CLI Options

The generator also supports command-line options that can be used with go:generate:

- `-dir` (string): Directory to scan for builder annotations (default: ".")
- `-prefix` (string): Default prefix for builder methods (default: "With")
- `-output` (string): Default output file pattern (default: "{name}_builder.go")
- `-package` (string): Default package name override
- `-validate` (bool): Enable validation by default

Note: Annotations in source files take precedence over CLI options unless the CLI options are explicitly set to non-default values.