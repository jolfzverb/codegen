# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

OpenAPI 3.0 code generator that creates type-safe Go code (HTTP handlers, models, validation) for REST APIs using Chi router and go-playground/validator.

**All comments and documentation must be written in English only.**

## Common Commands

```bash
make check      # Run all checks (generate, test, lint)
make test       # Run tests (go test ./...)
make lint       # Run linter (golangci-lint)
make generate   # Run go generate (generates example code)
```

Run a single test:
```bash
go test ./test -run TestName
go test ./internal/generator/astbuilder2 -run TestName
```

## Architecture

**Data Flow:** OpenAPI YAML → Generator → AST → Go Code

**Generation pipeline** (in `api.go:Generate()`): loops over YAML files, each goes through PrepareFiles() → GenerateFiles() → WriteOutFiles().

### Core Files

- `cmd/generate.go` - CLI entry point
- `internal/generator/api.go` - Main Generator struct and orchestration
- `internal/generator/generatehandlers.go` - Handler generation logic
- `internal/generator/generateschemas.go` - Schema generation logic
- `internal/generator/handlers.go`, `handlers2.go` - Handler AST building
- `internal/generator/schemas.go` - Schema AST building
- `internal/generator/utils.go` - AST helper functions: `I()`, `Str()`, `Field()`, `Func()`, `Sel()`
- `internal/generator/nameutils.go` - `FormatGoLikeIdentifier()`, `GoIdentLowercase()`

### AST Builders

- `internal/generator/astbuilder/` - Active implementation, being extended

### Testing

- Golden file tests in `test/testdata/` compare generated code against expected output
- Test OpenAPI specs in `test/yamls/`
- Generated example code in `internal/usage/generated/`

## Key Patterns

- Use `errors.Wrap()` for error handling
- Use `AddSchemasImport()` / `AddHandlersImport()` for managing imports in generated code
- Use `FormatGoLikeIdentifier()` to convert OpenAPI names to Go identifiers

## CLI Options

- `-d, --dir-prefix` - Output directory prefix
- `-p, --package-prefix` - Package import prefix
- `--pointers` - Generate required fields as pointers
- `--allow-delete-with-body` - Allow DELETE with request body
- `--allow-remote-addr-param` - Allow RemoteAddr parameter

## Common Pitfalls

- **AST structure**: Incorrect AST causes compilation errors in generated code — use utilities in `utils.go` and follow existing patterns
- **External references**: `$ref` to external files must be added to `YAMLFilesToProcess` via `ParseRefTypeName()`
- **Validation tags**: Check `SchemaField.TagValidate` population in schema processing
- **Imports**: Use `AddSchemasImport()` / `AddHandlersImport()` — missing imports are a common issue
