# Raw AST Usages to Migrate to Builder Pattern

All locations in `internal/generator/` where raw `go/ast` structs or bypass helpers are used instead of `astbuilder`.
Each group describes what new builder capability is needed and lists every occurrence.

---

## 1. `&ast.BinaryExpr{}` complex operators — extend binary expression builders

`Ne()` and `Eq()` exist in `ast_helpers.go` for `!=` and `==`.
Missing: `||` (LOR), `&&` (LAND), `+` (ADD) operators.
Should add `astbuilder.Or(x, y)`, `astbuilder.And(x, y)`, `astbuilder.Add(x, y)` or a general `BinaryExpr(op, x, y)`.

- [handler_ast.go:338-341](../../internal/generator/handler_ast.go#L338) — `&ast.BinaryExpr{X: Ne(...), Op: token.LOR, Y: Eq(...)}` — OR of two comparisons
- [parse_ast.go:216-220](../../internal/generator/parse_ast.go#L216) — `&ast.BinaryExpr{X: Ne(...), Op: token.LAND, Y: &ast.UnaryExpr{Op: token.NOT, ...}}` — AND NOT
- [validation_ast.go:177](../../internal/generator/validation_ast.go#L177) — `&ast.BinaryExpr{X: I("temp"), Op: token.EQL, Y: I("nil")}` — can use `Eq()` already
- [validation_ast.go:293-295](../../internal/generator/validation_ast.go#L293) — `&ast.BinaryExpr{X: &ast.BinaryExpr{X: Str("field "), Op: token.ADD, Y: I("field")}, Op: token.ADD, Y: Str(" is required")}` — string concat chain
- [validation_ast.go:299-302](../../internal/generator/validation_ast.go#L299) — `&ast.BinaryExpr{X: &ast.UnaryExpr{NOT, ...}, Op: token.LAND, Y: callExpr}` — NOT AND
- [validation_ast.go:306-308](../../internal/generator/validation_ast.go#L306) — string concat (same as line 293)
- [validation_ast.go:336-339](../../internal/generator/validation_ast.go#L336) — `&ast.BinaryExpr{X: I("exists"), Op: token.LAND, Y: &ast.UnaryExpr{NOT, ...}}`

---

## 2. `&ast.UnaryExpr{}` — add `astbuilder.Not(x)` helper

Unary NOT expressions used as conditions.

- [parse_ast.go:219-224](../../internal/generator/parse_ast.go#L219) — `&ast.UnaryExpr{Op: token.NOT, X: &ast.CallExpr{...errors.Is...}}`
- [validation_ast.go:290](../../internal/generator/validation_ast.go#L290) — `&ast.UnaryExpr{Op: token.NOT, X: I("exists")}`
- [validation_ast.go:300](../../internal/generator/validation_ast.go#L300) — `&ast.UnaryExpr{Op: token.NOT, X: &ast.IndexExpr{...}}`
- [validation_ast.go:339](../../internal/generator/validation_ast.go#L339) — `&ast.UnaryExpr{Op: token.NOT, X: &ast.CallExpr{...}}`
- [validation_ast.go:381-383](../../internal/generator/validation_ast.go#L381) — `&ast.UnaryExpr{Op: token.NOT, X: &ast.CallExpr{...containsNull...}}`

---

## 3. `&ast.DeclStmt{Decl: &ast.GenDecl{...ValueSpec{}}}` — extend `DeclareVar` builder

`astbuilder.DeclareVar(name, typeExpr)` exists but only handles simple `ast.Expr`.
All the places below pass raw types not yet covered (selector types, map types, array types).
Needs either extending `DeclareVar` or adding `DeclareVarWithType(name string, tb TypeExpressionBuilder)`.

- [handler_ast.go:470-477](../../internal/generator/handler_ast.go#L470) — `var headers map[string]string`
- [handler_ast.go:565-572](../../internal/generator/handler_ast.go#L565) — `var pathParams models.XxxPathParams`
- [parse_ast.go:16-24](../../internal/generator/parse_ast.go#L16) — `var queryParams models.XxxQueryParams`
- [parse_ast.go:119-127](../../internal/generator/parse_ast.go#L119) — `var headers models.XxxHeaders`
- [parse_ast.go:193-201](../../internal/generator/parse_ast.go#L193) — `var cookies models.XxxCookies`
- [parse_ast.go:299-307](../../internal/generator/parse_ast.go#L299) — `var bodyJSON json.RawMessage`
- [parse_ast.go:324-329](../../internal/generator/parse_ast.go#L324) — `var body <dynamic type>`
- [validation_ast.go:163-169](../../internal/generator/validation_ast.go#L163) — `var temp any`
- [validation_ast.go:256-263](../../internal/generator/validation_ast.go#L256) — `var obj map[string]json.RawMessage`
- [validation_ast.go:273-277](../../internal/generator/validation_ast.go#L273) — `var val json.RawMessage`
- [validation_ast.go:279-283](../../internal/generator/validation_ast.go#L279) — `var exists bool`
- [validation_ast.go:387-394](../../internal/generator/validation_ast.go#L387) — `var arr []json.RawMessage`

---

## 4. `&ast.KeyValueExpr{}` in composite literals — add `astbuilder.KeyValue(k, v)` helper

Used to build struct/map literal fields.

- [handler_ast.go:48-57](../../internal/generator/handler_ast.go#L48) — `InitHandlerConstructor`: `{Key: I("validator"), Value: &ast.CallExpr{...}}`
- [handler_ast.go:152-155](../../internal/generator/handler_ast.go#L152) — `AddDependencyToHandlers`: `{Key: I(fieldName), Value: I(fieldName)}`
- [parse_ast.go:375](../../internal/generator/parse_ast.go#L375) — `{Key: I("Path"), Value: Star(...)}`
- [parse_ast.go:381](../../internal/generator/parse_ast.go#L381) — `{Key: I("Query"), ...}`
- [parse_ast.go:387](../../internal/generator/parse_ast.go#L387) — `{Key: I("Headers"), ...}`
- [parse_ast.go:393](../../internal/generator/parse_ast.go#L393) — `{Key: I("Cookies"), ...}`
- [parse_ast.go:402](../../internal/generator/parse_ast.go#L402) — `{Key: I("Body"), Value: Star(...)}`
- [parse_ast.go:404](../../internal/generator/parse_ast.go#L404) — `{Key: I("Body"), Value: I("body")}`
- [response_ast.go:43](../../internal/generator/response_ast.go#L43) — `{Key: I("Body"), Value: I("body")}`
- [response_ast.go:49](../../internal/generator/response_ast.go#L49) — `{Key: I("Headers"), Value: I("headers")}`
- [response_ast.go:56-65](../../internal/generator/response_ast.go#L56) — `{Key: I("StatusCode"), ...}`, `{Key: I("Response"+code), Value: Amp(...)}`
- [validation_ast.go:238](../../internal/generator/validation_ast.go#L238) — `{Key: Str(fieldName), Value: I("true")}`
- [validation_ast.go:247](../../internal/generator/validation_ast.go#L247) — same for nullableFields

---

## 5. `&ast.CompositeLit{}` — add `astbuilder.CompositeLit(type, fields...)` builder

Used for struct and map literal expressions.

- [handler_ast.go:45-63](../../internal/generator/handler_ast.go#L45) — `InitHandlerConstructor`: `&ast.CompositeLit{Type: I("Handler"), Elts: [...]}`
- [parse_ast.go:413-415](../../internal/generator/parse_ast.go#L413) — `Amp(&ast.CompositeLit{Type: Sel(..., "XxxRequest"), Elts: elts})`
- [response_ast.go:53-68](../../internal/generator/response_ast.go#L53) — nested composite literals for `XxxResponse{StatusCode: ..., ResponseNNN: &XxxResponseNNN{...}}`
- [validation_ast.go:240-242](../../internal/generator/validation_ast.go#L240) — `&ast.CompositeLit{Type: &ast.MapType{...}, Elts: requiredFieldsElts}`
- [validation_ast.go:249-251](../../internal/generator/validation_ast.go#L249) — same for nullableFields map

---

## 6. `&ast.IndexExpr{}` — add `astbuilder.Index(x, index)` expression builder

Map/slice index expressions used as values.

- [validation_ast.go:289](../../internal/generator/validation_ast.go#L289) — `&ast.IndexExpr{X: I("obj"), Index: I("field")}`
- [validation_ast.go:300](../../internal/generator/validation_ast.go#L300) — `&ast.IndexExpr{X: I("nullableFields"), Index: I("field")}`
- [validation_ast.go:326](../../internal/generator/validation_ast.go#L326) — `&ast.IndexExpr{X: I("obj"), Index: Str(fieldName)}`

---

## 7. `&ast.MapType{}` / `&ast.ArrayType{}` — extend type builders

Raw map/array type expressions in variable declarations and composite literals.
`astbuilder.ArrayTypeBuilder` exists for `[]T` but not for `map[K]V`.

- [handler_ast.go:476](../../internal/generator/handler_ast.go#L476) — `&ast.MapType{Key: I("string"), Value: I("string")}` — `map[string]string`
- [validation_ast.go:241](../../internal/generator/validation_ast.go#L241) — `&ast.MapType{Key: I("string"), Value: I("bool")}` — `map[string]bool`
- [validation_ast.go:250](../../internal/generator/validation_ast.go#L250) — same map type
- [validation_ast.go:262](../../internal/generator/validation_ast.go#L262) — `&ast.MapType{Key: I("string"), Value: Sel(I("json"), "RawMessage")}` — `map[string]json.RawMessage`
- [validation_ast.go:393](../../internal/generator/validation_ast.go#L393) — `&ast.ArrayType{Elt: Sel(I("json"), "RawMessage")}` — `[]json.RawMessage`

---

## 8. `&ast.BasicLit{Kind: token.INT}` — add `IntLit(value string)` to `ast_helpers.go`

Integer literal expressions.

- [handler_ast.go:415](../../internal/generator/handler_ast.go#L415) — `&ast.BasicLit{Kind: token.INT, Value: code}` in switch case
- [response_ast.go:58](../../internal/generator/response_ast.go#L58) — `&ast.BasicLit{Kind: token.INT, Value: code}` in composite literal

---

## 9. `Ret1()` / `Ret2()` / `Ret()` helpers — replace with `astbuilder.Return*()`

These three helpers in `ast_helpers.go` duplicate what `astbuilder.Return1()`, `astbuilder.Return2()`, `astbuilder.Return()` already provide.

- [handler_ast.go:66](../../internal/generator/handler_ast.go#L66) — `Ret1(Amp(initializerComposite))` in `InitHandlerConstructor`

---

## 10. `Ne()` / `Eq()` helpers — replace with `astbuilder` equivalents or consolidate

`Ne()` and `Eq()` are defined in `ast_helpers.go`. Since `astbuilder` already has comparison helpers,
these could either be removed or re-exported from there to avoid dual sources.

Used throughout `handler_ast.go` and `parse_ast.go` in `If()` conditions — too many sites to list individually,
but all follow the pattern `astbuilder.If(Ne(I("err"), I("nil")))`.

---

## 11. `Star()` / `Amp()` / `Sel()` helpers — consolidate into `astbuilder`

These three shorthand functions from `ast_helpers.go` are fundamental building blocks used everywhere.
They could be promoted into `astbuilder` as package-level functions so that `ast_helpers.go`
(and its `go/ast` import) can eventually be removed from the generator layer.

Used pervasively across all `*_ast.go` files — no single location to pin.

---

## 12. `&ast.BlockStmt{}` empty initial block — replace with `astbuilder.NewBodyBuilder().Build()`

- [handler_ast.go:215-217](../../internal/generator/handler_ast.go#L215) — `switchBody := &ast.BlockStmt{List: []ast.Stmt{}}` in `CreateHandler`

---

## 13. `&ast.File{}` / top-level `&ast.GenDecl{Tok: token.IMPORT}` — consider a `FileBuilder`

These are at the file-output level and may warrant a dedicated `FileBuilder` or remain as is
if deemed infrastructural rather than generated code.

- [schema_ast.go:51-65](../../internal/generator/schema_ast.go#L51) — `WriteSchemasToOutput`: `&ast.File{Name: ..., Imports: ..., Decls: []}` + appended import `&ast.GenDecl{Tok: token.IMPORT}`
- [handler_ast.go:168-177](../../internal/generator/handler_ast.go#L168) — `GenerateHandlersFile`: same pattern

---

## Summary by file

| File | Raw AST count (approx) | Primary gap |
|---|---|---|
| [handler_ast.go](../../internal/generator/handler_ast.go) | ~30 | `&ast.CallExpr` inline, `DeclareVar` map type, `BinaryExpr` LOR |
| [parse_ast.go](../../internal/generator/parse_ast.go) | ~28 | `&ast.CallExpr` inline, `DeclareVar` selector types, `BinaryExpr` LAND NOT |
| [validation_ast.go](../../internal/generator/validation_ast.go) | ~30 | `DeclareVar` map/array types, `IndexExpr`, `BinaryExpr` ADD/LAND/NOT, `CompositeLit` |
| [response_ast.go](../../internal/generator/response_ast.go) | ~7 | `CompositeLit`, `KeyValueExpr`, `BasicLit` INT |
| [schema_ast.go](../../internal/generator/schema_ast.go) | 3 | `&ast.File{}`, import GenDecl |
| [ast_helpers.go](../../internal/generator/ast_helpers.go) | all | is itself raw AST — to be eliminated as each site migrates |
