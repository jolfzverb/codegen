package astbuilder

import (
	"go/ast"
	"go/format"
	"go/token"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func formatStmt(t *testing.T, stmt ast.Stmt) string {
	t.Helper()
	fset := token.NewFileSet()
	var buf strings.Builder
	err := format.Node(&buf, fset, stmt)
	require.NoError(t, err)
	return buf.String()
}

// ReturnBuilder tests

func TestReturnBuilder_Empty(t *testing.T) {
	stmt := Return().Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "return", result)
}

func TestReturnBuilder_SingleValue(t *testing.T) {
	stmt := Return1(ast.NewIdent("result")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "return result", result)
}

func TestReturnBuilder_TwoValues(t *testing.T) {
	stmt := Return2(ast.NewIdent("result"), ast.NewIdent("err")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "return result, err", result)
}

func TestReturnBuilder_Nil(t *testing.T) {
	stmt := ReturnNil().Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "return nil", result)
}

func TestReturnBuilder_Err(t *testing.T) {
	stmt := ReturnErr().Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "return err", result)
}

func TestReturnBuilder_NilErr(t *testing.T) {
	stmt := ReturnNilErr().Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "return nil, err", result)
}

func TestReturnBuilder_Clone(t *testing.T) {
	original := Return1(ast.NewIdent("a"))
	clone := original.Clone()
	clone.AddResult(ast.NewIdent("b"))

	assert.Equal(t, "return a", formatStmt(t, original.Build()))
	assert.Equal(t, "return a, b", formatStmt(t, clone.Build()))
}

// AssignBuilder tests

func TestAssignBuilder_Simple(t *testing.T) {
	stmt := Assign(ast.NewIdent("x"), ast.NewIdent("y")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "x = y", result)
}

func TestAssignBuilder_Define(t *testing.T) {
	stmt := Define(ast.NewIdent("x"), ast.NewIdent("y")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "x := y", result)
}

func TestAssignBuilder_Define2(t *testing.T) {
	stmt := Define2(
		ast.NewIdent("result"),
		ast.NewIdent("err"),
		&ast.CallExpr{Fun: ast.NewIdent("doWork")},
	).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "result, err := doWork()", result)
}

func TestAssignBuilder_DefineCall(t *testing.T) {
	stmt := DefineCall("result", ast.NewIdent("getValue")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "result := getValue()", result)
}

func TestAssignBuilder_DefineCallWithErr(t *testing.T) {
	stmt := DefineCallWithErr("result", ast.NewIdent("getValue")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "result, err := getValue()", result)
}

func TestAssignBuilder_Clone(t *testing.T) {
	original := Assign(ast.NewIdent("x"), ast.NewIdent("y"))
	clone := original.Clone()

	assert.Equal(t, "x = y", formatStmt(t, original.Build()))
	assert.Equal(t, "x = y", formatStmt(t, clone.Build()))
}

func TestAssignBuilder_PanicsOnEmptyLhs(t *testing.T) {
	assert.Panics(t, func() {
		NewAssignBuilder().AddRhs(ast.NewIdent("x")).Build()
	})
}

func TestAssignBuilder_PanicsOnEmptyRhs(t *testing.T) {
	assert.Panics(t, func() {
		NewAssignBuilder().AddLhs(ast.NewIdent("x")).Build()
	})
}

// VarDeclBuilder tests

func TestVarDeclBuilder_Simple(t *testing.T) {
	stmt := DeclareVar("x", ast.NewIdent("int")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "var x int", result)
}

func TestVarDeclBuilder_WithValue(t *testing.T) {
	stmt := DeclareVarWithValue("x", ast.NewIdent("int"), &ast.BasicLit{Kind: token.INT, Value: "42"}).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "var x int = 42", result)
}

func TestVarDeclBuilder_WithTypeBuilder(t *testing.T) {
	stmt := NewVarDeclBuilder().
		WithName("items").
		WithTypeBuilder(StringSlice()).
		Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "var items []string", result)
}

func TestVarDeclBuilder_Clone(t *testing.T) {
	original := DeclareVar("x", ast.NewIdent("int"))
	clone := original.Clone()

	assert.Equal(t, "var x int", formatStmt(t, original.Build()))
	assert.Equal(t, "var x int", formatStmt(t, clone.Build()))
}

func TestVarDeclBuilder_PanicsOnMissingName(t *testing.T) {
	assert.Panics(t, func() {
		NewVarDeclBuilder().WithType(ast.NewIdent("int")).Build()
	})
}

func TestVarDeclBuilder_PanicsOnMissingTypeAndValue(t *testing.T) {
	assert.Panics(t, func() {
		NewVarDeclBuilder().WithName("x").Build()
	})
}

// IfBuilder tests

func TestIfBuilder_Simple(t *testing.T) {
	stmt := If(&ast.BinaryExpr{
		X:  ast.NewIdent("x"),
		Op: token.GTR,
		Y:  &ast.BasicLit{Kind: token.INT, Value: "0"},
	}).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "if x > 0 {\n}", result)
}

func TestIfBuilder_WithBody(t *testing.T) {
	stmt := If(&ast.BinaryExpr{
		X:  ast.NewIdent("x"),
		Op: token.GTR,
		Y:  &ast.BasicLit{Kind: token.INT, Value: "0"},
	}).WithBody(NewBodyBuilder().Return1(ast.NewIdent("x"))).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "if x > 0 {\n\treturn x\n}", result)
}

func TestIfBuilder_WithElse(t *testing.T) {
	stmt := If(&ast.BinaryExpr{
		X:  ast.NewIdent("x"),
		Op: token.GTR,
		Y:  &ast.BasicLit{Kind: token.INT, Value: "0"},
	}).
		WithBody(NewBodyBuilder().Return1(ast.NewIdent("x"))).
		WithElse(NewBodyBuilder().Return1(&ast.UnaryExpr{Op: token.SUB, X: ast.NewIdent("x")})).
		Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "if x > 0 {\n\treturn x\n} else {\n\treturn -x\n}", result)
}

func TestIfBuilder_ErrNotNil(t *testing.T) {
	stmt := IfErrNotNil().Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "if err != nil {\n}", result)
}

func TestIfBuilder_ErrNotNilReturn(t *testing.T) {
	stmt := IfErrNotNilReturn(ast.NewIdent("nil")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "if err != nil {\n\treturn nil, err\n}", result)
}

func TestIfBuilder_Nil(t *testing.T) {
	stmt := IfNil(ast.NewIdent("result")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "if result == nil {\n}", result)
}

func TestIfBuilder_NotNil(t *testing.T) {
	stmt := IfNotNil(ast.NewIdent("result")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "if result != nil {\n}", result)
}

func TestIfBuilder_WithInit(t *testing.T) {
	stmt := NewIfBuilder().
		WithInitBuilder(Define(ast.NewIdent("x"), &ast.CallExpr{Fun: ast.NewIdent("getValue")})).
		WithCond(&ast.BinaryExpr{X: ast.NewIdent("x"), Op: token.NEQ, Y: ast.NewIdent("nil")}).
		Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "if x := getValue(); x != nil {\n}", result)
}

func TestIfBuilder_DirectBodyManipulation(t *testing.T) {
	ib := If(&ast.BinaryExpr{X: ast.NewIdent("x"), Op: token.GTR, Y: &ast.BasicLit{Kind: token.INT, Value: "0"}})
	ib.Body().Return1(ast.NewIdent("x"))
	result := formatStmt(t, ib.Build())
	assert.Equal(t, "if x > 0 {\n\treturn x\n}", result)
}

func TestIfBuilder_Clone(t *testing.T) {
	original := IfErrNotNil().WithBody(NewBodyBuilder().Return1(ast.NewIdent("err")))
	clone := original.Clone()

	assert.Equal(t, "if err != nil {\n\treturn err\n}", formatStmt(t, original.Build()))
	assert.Equal(t, "if err != nil {\n\treturn err\n}", formatStmt(t, clone.Build()))
}

func TestIfBuilder_PanicsOnMissingCond(t *testing.T) {
	assert.Panics(t, func() {
		NewIfBuilder().Build()
	})
}

// ExprStmtBuilder tests

func TestExprStmtBuilder_Simple(t *testing.T) {
	stmt := ExprStmt(&ast.CallExpr{Fun: ast.NewIdent("doWork")}).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "doWork()", result)
}

func TestExprStmtBuilder_CallStmt(t *testing.T) {
	stmt := CallStmt(ast.NewIdent("print"), &ast.BasicLit{Kind: token.STRING, Value: `"hello"`}).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, `print("hello")`, result)
}

func TestExprStmtBuilder_MethodCallStmt(t *testing.T) {
	stmt := MethodCallStmt("fmt", "Println", &ast.BasicLit{Kind: token.STRING, Value: `"hello"`}).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, `fmt.Println("hello")`, result)
}

func TestExprStmtBuilder_Clone(t *testing.T) {
	original := CallStmt(ast.NewIdent("foo"))
	clone := original.Clone()

	assert.Equal(t, "foo()", formatStmt(t, original.Build()))
	assert.Equal(t, "foo()", formatStmt(t, clone.Build()))
}

func TestExprStmtBuilder_PanicsOnMissingExpr(t *testing.T) {
	assert.Panics(t, func() {
		NewExprStmtBuilder().Build()
	})
}

// RangeBuilder tests

func TestRangeBuilder_Simple(t *testing.T) {
	stmt := Range("i", "v", ast.NewIdent("items")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "for i, v := range items {\n}", result)
}

func TestRangeBuilder_ValueOnly(t *testing.T) {
	stmt := RangeValue("item", ast.NewIdent("items")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "for _, item := range items {\n}", result)
}

func TestRangeBuilder_IndexOnly(t *testing.T) {
	stmt := RangeIndex("i", ast.NewIdent("items")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "for i := range items {\n}", result)
}

func TestRangeBuilder_WithBody(t *testing.T) {
	stmt := RangeValue("item", ast.NewIdent("items")).
		WithBody(NewBodyBuilder().Call(ast.NewIdent("process"), ast.NewIdent("item"))).
		Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "for _, item := range items {\n\tprocess(item)\n}", result)
}

func TestRangeBuilder_DirectBodyManipulation(t *testing.T) {
	rb := RangeValue("item", ast.NewIdent("items"))
	rb.Body().Call(ast.NewIdent("process"), ast.NewIdent("item"))
	result := formatStmt(t, rb.Build())
	assert.Equal(t, "for _, item := range items {\n\tprocess(item)\n}", result)
}

func TestRangeBuilder_AsAssign(t *testing.T) {
	stmt := Range("i", "v", ast.NewIdent("items")).AsAssign().Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "for i, v = range items {\n}", result)
}

func TestRangeBuilder_Clone(t *testing.T) {
	original := RangeValue("item", ast.NewIdent("items"))
	clone := original.Clone()

	assert.Equal(t, "for _, item := range items {\n}", formatStmt(t, original.Build()))
	assert.Equal(t, "for _, item := range items {\n}", formatStmt(t, clone.Build()))
}

func TestRangeBuilder_PanicsOnMissingX(t *testing.T) {
	assert.Panics(t, func() {
		NewRangeBuilder().WithKey(ast.NewIdent("i")).Build()
	})
}

// SwitchBuilder tests

func TestSwitchBuilder_Simple(t *testing.T) {
	stmt := Switch(ast.NewIdent("x")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "switch x {\n}", result)
}

func TestSwitchBuilder_WithCases(t *testing.T) {
	stmt := Switch(ast.NewIdent("x")).
		AddCase(Case(&ast.BasicLit{Kind: token.INT, Value: "1"}).
			WithBody(NewBodyBuilder().Return1(&ast.BasicLit{Kind: token.STRING, Value: `"one"`}))).
		AddCase(Case(&ast.BasicLit{Kind: token.INT, Value: "2"}).
			WithBody(NewBodyBuilder().Return1(&ast.BasicLit{Kind: token.STRING, Value: `"two"`}))).
		AddCase(Default().
			WithBody(NewBodyBuilder().Return1(&ast.BasicLit{Kind: token.STRING, Value: `"other"`}))).
		Build()
	result := formatStmt(t, stmt)
	expected := "switch x {\ncase 1:\n\treturn \"one\"\ncase 2:\n\treturn \"two\"\ndefault:\n\treturn \"other\"\n}"
	assert.Equal(t, expected, result)
}

func TestSwitchBuilder_CaseWithMultipleExprs(t *testing.T) {
	stmt := Switch(ast.NewIdent("x")).
		AddCase(Case(&ast.BasicLit{Kind: token.INT, Value: "1"}, &ast.BasicLit{Kind: token.INT, Value: "2"}).
			WithBody(NewBodyBuilder().Return1(&ast.BasicLit{Kind: token.STRING, Value: `"small"`}))).
		Build()
	result := formatStmt(t, stmt)
	expected := "switch x {\ncase 1, 2:\n\treturn \"small\"\n}"
	assert.Equal(t, expected, result)
}

func TestSwitchBuilder_Clone(t *testing.T) {
	original := Switch(ast.NewIdent("x")).
		AddCase(Case(&ast.BasicLit{Kind: token.INT, Value: "1"}))
	clone := original.Clone()

	assert.Equal(t, "switch x {\ncase 1:\n}", formatStmt(t, original.Build()))
	assert.Equal(t, "switch x {\ncase 1:\n}", formatStmt(t, clone.Build()))
}

// Integration tests

func TestBodyBuilder_WithStatementBuilders(t *testing.T) {
	body := NewBodyBuilder().
		AddStmt(DeclareVar("result", ast.NewIdent("string"))).
		AddStmt(Define(ast.NewIdent("err"), &ast.CallExpr{Fun: ast.NewIdent("doWork")})).
		AddStmt(IfErrNotNilReturn(ast.NewIdent("result"))).
		AddStmt(Return1(ast.NewIdent("result"))).
		Build()

	fset := token.NewFileSet()
	var buf strings.Builder
	err := format.Node(&buf, fset, body)
	require.NoError(t, err)

	expected := "{\n\tvar result string\n\terr := doWork()\n\tif err != nil {\n\t\treturn result, err\n\t}\n\treturn result\n}"
	assert.Equal(t, expected, buf.String())
}

func TestFunctionBuilder_WithStatementBuilders(t *testing.T) {
	fn := Method("h", "Handler", "Process").
		AddParam(StringField("input")).
		AddResult(StringField("")).
		AddErrorResult()

	fn.Body().
		AddStmt(DeclareVar("result", ast.NewIdent("string"))).
		AddStmt(DefineCallWithErr("data", &ast.SelectorExpr{X: ast.NewIdent("h"), Sel: ast.NewIdent("fetch")}, ast.NewIdent("input"))).
		AddStmt(IfErrNotNilReturn(&ast.BasicLit{Kind: token.STRING, Value: `""`})).
		AddStmt(Return2(ast.NewIdent("data"), ast.NewIdent("nil")))

	fset := token.NewFileSet()
	var buf strings.Builder
	err := format.Node(&buf, fset, fn.Build())
	require.NoError(t, err)

	expected := `func (h *Handler) Process(input string) (string, error) {
	var result string
	data, err := h.fetch(input)
	if err != nil {
		return "", err
	}
	return data, nil
}`
	assert.Equal(t, expected, buf.String())
}
