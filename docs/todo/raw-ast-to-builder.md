# Raw AST Usages to Migrate to Builder Pattern

All locations in `internal/generator/` where raw `go/ast` structs or bypass helpers are used instead of `astbuilder`.
Each group describes what new builder capability is needed and lists every occurrence.

---

## 1. `&ast.CompositeLit{}` — add `astbuilder.CompositeLit(type, fields...)` builder

Used for struct and map literal expressions.

- [handler_ast.go:45-63](../../internal/generator/handler_ast.go#L45) — `InitHandlerConstructor`: `&ast.CompositeLit{Type: I("Handler"), Elts: [...]}`
- [parse_ast.go:413-415](../../internal/generator/parse_ast.go#L413) — `Amp(&ast.CompositeLit{Type: Sel(..., "XxxRequest"), Elts: elts})`
- [response_ast.go:53-68](../../internal/generator/response_ast.go#L53) — nested composite literals for `XxxResponse{StatusCode: ..., ResponseNNN: &XxxResponseNNN{...}}`
- [validation_ast.go:240-242](../../internal/generator/validation_ast.go#L240) — `&ast.CompositeLit{Type: &ast.MapType{...}, Elts: requiredFieldsElts}`
- [validation_ast.go:249-251](../../internal/generator/validation_ast.go#L249) — same for nullableFields map

---

## 2. `&ast.IndexExpr{}` — add `astbuilder.Index(x, index)` expression builder

Map/slice index expressions used as values.

- [validation_ast.go:289](../../internal/generator/validation_ast.go#L289) — `&ast.IndexExpr{X: I("obj"), Index: I("field")}`
- [validation_ast.go:300](../../internal/generator/validation_ast.go#L300) — `&ast.IndexExpr{X: I("nullableFields"), Index: I("field")}`
- [validation_ast.go:326](../../internal/generator/validation_ast.go#L326) — `&ast.IndexExpr{X: I("obj"), Index: Str(fieldName)}`

---

## 3. `&ast.MapType{}` / `&ast.ArrayType{}` — extend type builders

Raw map/array type expressions in variable declarations and composite literals.
`astbuilder.ArrayTypeBuilder` exists for `[]T` but not for `map[K]V`.

- [handler_ast.go:476](../../internal/generator/handler_ast.go#L476) — `&ast.MapType{Key: I("string"), Value: I("string")}` — `map[string]string`
- [validation_ast.go:241](../../internal/generator/validation_ast.go#L241) — `&ast.MapType{Key: I("string"), Value: I("bool")}` — `map[string]bool`
- [validation_ast.go:250](../../internal/generator/validation_ast.go#L250) — same map type
- [validation_ast.go:262](../../internal/generator/validation_ast.go#L262) — `&ast.MapType{Key: I("string"), Value: Sel(I("json"), "RawMessage")}` — `map[string]json.RawMessage`
- [validation_ast.go:393](../../internal/generator/validation_ast.go#L393) — `&ast.ArrayType{Elt: Sel(I("json"), "RawMessage")}` — `[]json.RawMessage`

---

## 4. `&ast.BasicLit{Kind: token.INT}` — add `IntLit(value string)` to `ast_helpers.go`

Integer literal expressions.

- [handler_ast.go:415](../../internal/generator/handler_ast.go#L415) — `&ast.BasicLit{Kind: token.INT, Value: code}` in switch case
- [response_ast.go:58](../../internal/generator/response_ast.go#L58) — `&ast.BasicLit{Kind: token.INT, Value: code}` in composite literal

---

## 5. `Ret1()` / `Ret2()` / `Ret()` helpers — replace with `astbuilder.Return*()`

These three helpers in `ast_helpers.go` duplicate what `astbuilder.Return1()`, `astbuilder.Return2()`, `astbuilder.Return()` already provide.

- [handler_ast.go:66](../../internal/generator/handler_ast.go#L66) — `Ret1(Amp(initializerComposite))` in `InitHandlerConstructor`

---

## 6. `Ne()` / `Eq()` helpers — replace with `astbuilder` equivalents or consolidate

`Ne()` and `Eq()` are defined in `ast_helpers.go`. Since `astbuilder` already has comparison helpers,
these could either be removed or re-exported from there to avoid dual sources.

Used throughout `handler_ast.go` and `parse_ast.go` in `If()` conditions — too many sites to list individually,
but all follow the pattern `astbuilder.If(Ne(I("err"), I("nil")))`.

---

## 7. `Star()` / `Amp()` / `Sel()` helpers — consolidate into `astbuilder`

These three shorthand functions from `ast_helpers.go` are fundamental building blocks used everywhere.
They could be promoted into `astbuilder` as package-level functions so that `ast_helpers.go`
(and its `go/ast` import) can eventually be removed from the generator layer.

Used pervasively across all `*_ast.go` files — no single location to pin.

---

## 8. `&ast.BlockStmt{}` empty initial block — replace with `astbuilder.NewBodyBuilder().Build()`

- [handler_ast.go:215-217](../../internal/generator/handler_ast.go#L215) — `switchBody := &ast.BlockStmt{List: []ast.Stmt{}}` in `CreateHandler`

---

## 9. `&ast.File{}` / top-level `&ast.GenDecl{Tok: token.IMPORT}` — consider a `FileBuilder`

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
