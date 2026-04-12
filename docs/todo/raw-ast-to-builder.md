# Raw AST Usages to Migrate to Builder Pattern

All locations in `internal/generator/` where raw `go/ast` structs or bypass helpers are used instead of `astbuilder`.
Each group describes what new builder capability is needed and lists every occurrence.

---

## 1. `&ast.BlockStmt{}` empty initial block — replace with `astbuilder.NewBodyBuilder().Build()`

- [handler_ast.go:215-217](../../internal/generator/handler_ast.go#L215) — `switchBody := &ast.BlockStmt{List: []ast.Stmt{}}` in `CreateHandler`

---

## 2. `&ast.File{}` / top-level `&ast.GenDecl{Tok: token.IMPORT}` — consider a `FileBuilder`

These are at the file-output level and may warrant a dedicated `FileBuilder` or remain as is
if deemed infrastructural rather than generated code.

- [schema_ast.go:51-65](../../internal/generator/schema_ast.go#L51) — `WriteSchemasToOutput`: `&ast.File{Name: ..., Imports: ..., Decls: []}` + appended import `&ast.GenDecl{Tok: token.IMPORT}`
- [handler_ast.go:168-177](../../internal/generator/handler_ast.go#L168) — `GenerateHandlersFile`: same pattern

---

## Summary by file

| File | Raw AST count (approx) | Primary gap |
|---|---|---|
| [handler_ast.go](../../internal/generator/handler_ast.go) | ~25 | `DeclareVar` map type, `BlockStmt`, `BasicLit` INT |
| [parse_ast.go](../../internal/generator/parse_ast.go) | ~20 | `DeclareVar` selector types, `KeyValueExpr`, `CompositeLit` |
| [validation_ast.go](../../internal/generator/validation_ast.go) | ~25 | `DeclareVar` map/array types, `IndexExpr`, `UnaryExpr`, `CompositeLit` |
| [response_ast.go](../../internal/generator/response_ast.go) | ~7 | `CompositeLit`, `KeyValueExpr`, `BasicLit` INT |
| [schema_ast.go](../../internal/generator/schema_ast.go) | 3 | `&ast.File{}`, import GenDecl |
| [ast_helpers.go](../../internal/generator/ast_helpers.go) | all | is itself raw AST — to be eliminated as each site migrates |
