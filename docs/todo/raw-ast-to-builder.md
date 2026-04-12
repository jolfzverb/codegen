# Raw AST Usages to Migrate to Builder Pattern

All locations in `internal/generator/` where raw `go/ast` structs or bypass helpers are used instead of `astbuilder`.
Each group describes what new builder capability is needed and lists every occurrence.

---

## 1. `Ne()` / `Eq()` helpers — replace with `astbuilder` equivalents or consolidate

`Ne()` and `Eq()` are defined in `ast_helpers.go`. Since `astbuilder` already has comparison helpers,
these could either be removed or re-exported from there to avoid dual sources.

Used throughout `handler_ast.go` and `parse_ast.go` in `If()` conditions — too many sites to list individually,
but all follow the pattern `astbuilder.If(Ne(I("err"), I("nil")))`.

---

## 2. `Star()` / `Amp()` / `Sel()` helpers — consolidate into `astbuilder`

These three shorthand functions from `ast_helpers.go` are fundamental building blocks used everywhere.
They could be promoted into `astbuilder` as package-level functions so that `ast_helpers.go`
(and its `go/ast` import) can eventually be removed from the generator layer.

Used pervasively across all `*_ast.go` files — no single location to pin.

---

## 3. `&ast.BlockStmt{}` empty initial block — replace with `astbuilder.NewBodyBuilder().Build()`

- [handler_ast.go:215-217](../../internal/generator/handler_ast.go#L215) — `switchBody := &ast.BlockStmt{List: []ast.Stmt{}}` in `CreateHandler`

---

## 4. `&ast.File{}` / top-level `&ast.GenDecl{Tok: token.IMPORT}` — consider a `FileBuilder`

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
