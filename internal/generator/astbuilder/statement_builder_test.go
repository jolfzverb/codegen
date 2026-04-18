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
	stmt := Return1(I("result")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "return result", result)
}

func TestReturnBuilder_TwoValues(t *testing.T) {
	stmt := Return2(I("result"), I("err")).Build()
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

// AssignBuilder tests

func TestAssignBuilder_Simple(t *testing.T) {
	stmt := Assign(I("x"), I("y")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "x = y", result)
}

func TestAssignBuilder_Define(t *testing.T) {
	stmt := Define(I("x"), I("y")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "x := y", result)
}

func TestAssignBuilder_Define2(t *testing.T) {
	stmt := Define2(
		I("result"),
		I("err"),
		Call(I("doWork")),
	).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "result, err := doWork()", result)
}

func TestAssignBuilder_DefineCall(t *testing.T) {
	stmt := DefineCall("result", I("getValue")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "result := getValue()", result)
}

func TestAssignBuilder_DefineCallWithErr(t *testing.T) {
	stmt := DefineCallWithErr("result", I("getValue")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "result, err := getValue()", result)
}

func TestAssignBuilder_PanicsOnEmptyLhs(t *testing.T) {
	assert.Panics(t, func() {
		NewAssignBuilder().AddRhs(I("x")).Build()
	})
}

func TestAssignBuilder_PanicsOnEmptyRhs(t *testing.T) {
	assert.Panics(t, func() {
		NewAssignBuilder().AddLhs(I("x")).Build()
	})
}

// VarDeclBuilder tests

func TestVarDeclBuilder_Simple(t *testing.T) {
	stmt := DeclareVar("x", I("int")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "var x int", result)
}

func TestVarDeclBuilder_WithValue(t *testing.T) {
	stmt := DeclareVarWithValue("x", I("int"), IntLit("42")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "var x int = 42", result)
}

func TestVarDeclBuilder_WithType(t *testing.T) {
	stmt := NewVarDeclBuilder().
		WithName("items").
		WithType(SliceOf(String())).
		Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "var items []string", result)
}

func TestVarDeclBuilder_PanicsOnMissingName(t *testing.T) {
	assert.Panics(t, func() {
		NewVarDeclBuilder().WithType(I("int")).Build()
	})
}

func TestVarDeclBuilder_PanicsOnMissingTypeAndValue(t *testing.T) {
	assert.Panics(t, func() {
		NewVarDeclBuilder().WithName("x").Build()
	})
}

// IfBuilder tests

func TestIfBuilder_Simple(t *testing.T) {
	stmt := If(Gt(I("x"), IntLit("0"))).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "if x > 0 {\n}", result)
}

func TestIfBuilder_WithBody(t *testing.T) {
	stmt := If(Gt(I("x"), IntLit("0"))).
		WithBody(NewBodyBuilder().Return1(I("x"))).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "if x > 0 {\n\treturn x\n}", result)
}

func TestIfBuilder_WithElse(t *testing.T) {
	stmt := If(Gt(I("x"), IntLit("0"))).
		WithBody(NewBodyBuilder().Return1(I("x"))).
		WithElse(NewBodyBuilder().Return1(Neg(I("x")))).
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
	stmt := IfErrNotNilReturn(I("nil")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "if err != nil {\n\treturn nil, err\n}", result)
}

func TestIfBuilder_Nil(t *testing.T) {
	stmt := IfNil(I("result")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "if result == nil {\n}", result)
}

func TestIfBuilder_NotNil(t *testing.T) {
	stmt := IfNotNil(I("result")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "if result != nil {\n}", result)
}

func TestIfBuilder_WithInit(t *testing.T) {
	stmt := NewIfBuilder().
		WithInit(Define(I("x"), Call(I("getValue")))).
		WithCond(Ne(I("x"), I("nil"))).
		Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "if x := getValue(); x != nil {\n}", result)
}

func TestIfBuilder_DirectBodyManipulation(t *testing.T) {
	ib := If(Gt(I("x"), IntLit("0")))
	ib.Body().Return1(I("x"))
	result := formatStmt(t, ib.Build())
	assert.Equal(t, "if x > 0 {\n\treturn x\n}", result)
}

func TestIfBuilder_PanicsOnMissingCond(t *testing.T) {
	assert.Panics(t, func() {
		NewIfBuilder().Build()
	})
}

// ExprStmtBuilder tests

func TestExprStmtBuilder_Simple(t *testing.T) {
	stmt := ExprStmt(Call(I("doWork"))).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "doWork()", result)
}

func TestExprStmtBuilder_CallStmt(t *testing.T) {
	stmt := CallStmt(I("print"), Str("hello")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, `print("hello")`, result)
}

func TestExprStmtBuilder_MethodCallStmt(t *testing.T) {
	stmt := MethodCallStmt("fmt", "Println", Str("hello")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, `fmt.Println("hello")`, result)
}

func TestExprStmtBuilder_PanicsOnMissingExpr(t *testing.T) {
	assert.Panics(t, func() {
		NewExprStmtBuilder().Build()
	})
}

// RangeBuilder tests

func TestRangeBuilder_Simple(t *testing.T) {
	stmt := Range("i", "v", I("items")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "for i, v := range items {\n}", result)
}

func TestRangeBuilder_ValueOnly(t *testing.T) {
	stmt := RangeValue("item", I("items")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "for _, item := range items {\n}", result)
}

func TestRangeBuilder_IndexOnly(t *testing.T) {
	stmt := RangeIndex("i", I("items")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "for i := range items {\n}", result)
}

func TestRangeBuilder_WithBody(t *testing.T) {
	stmt := RangeValue("item", I("items")).
		WithBody(NewBodyBuilder().Call(I("process"), I("item"))).
		Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "for _, item := range items {\n\tprocess(item)\n}", result)
}

func TestRangeBuilder_DirectBodyManipulation(t *testing.T) {
	rb := RangeValue("item", I("items"))
	rb.Body().Call(I("process"), I("item"))
	result := formatStmt(t, rb.Build())
	assert.Equal(t, "for _, item := range items {\n\tprocess(item)\n}", result)
}

func TestRangeBuilder_AsAssign(t *testing.T) {
	stmt := Range("i", "v", I("items")).AsAssign().Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "for i, v = range items {\n}", result)
}

func TestRangeBuilder_PanicsOnMissingX(t *testing.T) {
	assert.Panics(t, func() {
		NewRangeBuilder().WithKey(I("i")).Build()
	})
}

// SwitchBuilder tests

func TestSwitchBuilder_Simple(t *testing.T) {
	stmt := Switch(I("x")).Build()
	result := formatStmt(t, stmt)
	assert.Equal(t, "switch x {\n}", result)
}

func TestSwitchBuilder_WithCases(t *testing.T) {
	stmt := Switch(I("x")).
		AddCase(Case(IntLit("1")).
			WithBody(NewBodyBuilder().Return1(Str("one")))).
		AddCase(Case(IntLit("2")).
			WithBody(NewBodyBuilder().Return1(Str("two")))).
		AddCase(Default().
			WithBody(NewBodyBuilder().Return1(Str("other")))).
		Build()
	result := formatStmt(t, stmt)
	expected := "switch x {\ncase 1:\n\treturn \"one\"\ncase 2:\n\treturn \"two\"\ndefault:\n\treturn \"other\"\n}"
	assert.Equal(t, expected, result)
}

func TestSwitchBuilder_CaseWithMultipleExprs(t *testing.T) {
	stmt := Switch(I("x")).
		AddCase(Case(IntLit("1"), IntLit("2")).
			WithBody(NewBodyBuilder().Return1(Str("small")))).
		Build()
	result := formatStmt(t, stmt)
	expected := "switch x {\ncase 1, 2:\n\treturn \"small\"\n}"
	assert.Equal(t, expected, result)
}

// Integration tests

func TestBodyBuilder_WithStatementBuilders(t *testing.T) {
	body := NewBodyBuilder().
		AddStmt(DeclareVar("result", I("string"))).
		AddStmt(Define(I("err"), Call(I("doWork")))).
		AddStmt(IfErrNotNilReturn(I("result"))).
		AddStmt(Return1(I("result"))).
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
		AddStmt(DeclareVar("result", I("string"))).
		AddStmt(DefineCallWithErr("data", Sel(I("h"), "fetch"), I("input"))).
		AddStmt(IfErrNotNilReturn(Str(""))).
		AddStmt(Return2(I("data"), I("nil")))

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
