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

func formatFuncDecl(t *testing.T, funcDecl *ast.FuncDecl) string {
	t.Helper()
	fset := token.NewFileSet()
	var buf strings.Builder
	err := format.Node(&buf, fset, funcDecl)
	require.NoError(t, err)
	return buf.String()
}

func formatBlockStmt(t *testing.T, blockStmt *ast.BlockStmt) string {
	t.Helper()
	fset := token.NewFileSet()
	var buf strings.Builder
	err := format.Node(&buf, fset, blockStmt)
	require.NoError(t, err)
	return buf.String()
}

func TestFunctionBuilder_SimpleFunction(t *testing.T) {
	fn := NewFunctionBuilder().
		WithName("DoSomething").
		Build()

	result := formatFuncDecl(t, fn)
	assert.Equal(t, "func DoSomething() {\n}", result)
}

func TestFunctionBuilder_WithParams(t *testing.T) {
	fn := NewFunctionBuilder().
		WithName("Process").
		AddParam(StringField("name")).
		AddParam(IntField("count")).
		Build()

	result := formatFuncDecl(t, fn)
	assert.Equal(t, "func Process(name string, count int) {\n}", result)
}

func TestFunctionBuilder_WithResults(t *testing.T) {
	fn := NewFunctionBuilder().
		WithName("GetValue").
		AddResult(StringField("")).
		AddErrorResult().
		Build()

	result := formatFuncDecl(t, fn)
	assert.Equal(t, "func GetValue() (string, error) {\n}", result)
}

func TestFunctionBuilder_WithReceiver(t *testing.T) {
	fn := NewFunctionBuilder().
		WithName("Method").
		WithPointerReceiver("h", "Handler").
		Build()

	result := formatFuncDecl(t, fn)
	assert.Equal(t, "func (h *Handler) Method() {\n}", result)
}

func TestFunctionBuilder_WithValueReceiver(t *testing.T) {
	fn := NewFunctionBuilder().
		WithName("Method").
		WithValueReceiver("h", "Handler").
		Build()

	result := formatFuncDecl(t, fn)
	assert.Equal(t, "func (h Handler) Method() {\n}", result)
}

func TestFunctionBuilder_CompleteMethod(t *testing.T) {
	fn := NewFunctionBuilder().
		WithName("HandleRequest").
		WithPointerReceiver("h", "Handler").
		AddContextParam("ctx").
		AddParam(SelectorField("r", "models", "Request")).
		AddResult(NewFieldBuilder().WithType(Selector("models", "Response").AsPointer(true))).
		AddErrorResult().
		Build()

	result := formatFuncDecl(t, fn)
	expected := "func (h *Handler) HandleRequest(ctx context.Context, r models.Request) (*models.Response, error) {\n}"
	assert.Equal(t, expected, result)
}

func TestFunctionBuilder_HelperFunction(t *testing.T) {
	fn := Function("SimpleFunc").Build()

	result := formatFuncDecl(t, fn)
	assert.Equal(t, "func SimpleFunc() {\n}", result)
}

func TestFunctionBuilder_HelperMethod(t *testing.T) {
	fn := Method("s", "Service", "DoWork").Build()

	result := formatFuncDecl(t, fn)
	assert.Equal(t, "func (s *Service) DoWork() {\n}", result)
}

func TestFunctionBuilder_Clone(t *testing.T) {
	original := NewFunctionBuilder().
		WithName("Test").
		WithPointerReceiver("h", "Handler").
		AddParam(StringField("name"))

	clone := original.Clone()
	clone.WithName("TestClone")

	assert.Equal(t, "Test", original.GetName())
	assert.Equal(t, "TestClone", clone.GetName())
}

func TestFunctionBuilder_UtilityMethods(t *testing.T) {
	fb := NewFunctionBuilder().
		WithName("Test").
		WithPointerReceiver("h", "Handler").
		AddParam(StringField("name")).
		AddResult(ErrorField())

	assert.True(t, fb.HasName())
	assert.Equal(t, "Test", fb.GetName())
	assert.True(t, fb.HasReceiver())
	assert.Equal(t, 1, fb.ParamCount())
	assert.Equal(t, 1, fb.ResultCount())
}

func TestFunctionBuilder_PanicsOnMissingName(t *testing.T) {
	assert.Panics(t, func() {
		NewFunctionBuilder().Build()
	})
}

func TestBodyBuilder_Empty(t *testing.T) {
	body := NewBodyBuilder().Build()

	result := formatBlockStmt(t, body)
	assert.Equal(t, "{\n}", result)
}

func TestBodyBuilder_Return(t *testing.T) {
	body := NewBodyBuilder().Return().Build()

	result := formatBlockStmt(t, body)
	assert.Equal(t, "{\n\treturn\n}", result)
}

func TestBodyBuilder_Return1(t *testing.T) {
	body := NewBodyBuilder().
		Return1(ast.NewIdent("nil")).
		Build()

	result := formatBlockStmt(t, body)
	assert.Equal(t, "{\n\treturn nil\n}", result)
}

func TestBodyBuilder_Return2(t *testing.T) {
	body := NewBodyBuilder().
		Return2(ast.NewIdent("result"), ast.NewIdent("nil")).
		Build()

	result := formatBlockStmt(t, body)
	assert.Equal(t, "{\n\treturn result, nil\n}", result)
}

func TestBodyBuilder_DeclareVar(t *testing.T) {
	body := NewBodyBuilder().
		DeclareVar("err", ast.NewIdent("error")).
		Build()

	result := formatBlockStmt(t, body)
	assert.Equal(t, "{\n\tvar err error\n}", result)
}

func TestBodyBuilder_Define(t *testing.T) {
	body := NewBodyBuilder().
		AddStmt(Define2(
			ast.NewIdent("result"),
			ast.NewIdent("err"),
			&ast.CallExpr{Fun: ast.NewIdent("doWork")},
		)).
		Build()

	result := formatBlockStmt(t, body)
	assert.Equal(t, "{\n\tresult, err := doWork()\n}", result)
}

func TestBodyBuilder_If(t *testing.T) {
	body := NewBodyBuilder().
		If(
			&ast.BinaryExpr{X: ast.NewIdent("err"), Op: token.NEQ, Y: ast.NewIdent("nil")},
			NewBodyBuilder().Return1(ast.NewIdent("err")),
		).
		Build()

	result := formatBlockStmt(t, body)
	expected := "{\n\tif err != nil {\n\t\treturn err\n\t}\n}"
	assert.Equal(t, expected, result)
}

func TestBodyBuilder_IfElse(t *testing.T) {
	body := NewBodyBuilder().
		IfElse(
			&ast.BinaryExpr{X: ast.NewIdent("x"), Op: token.GTR, Y: &ast.BasicLit{Kind: token.INT, Value: "0"}},
			NewBodyBuilder().Return1(ast.NewIdent("x")),
			NewBodyBuilder().Return1(&ast.UnaryExpr{Op: token.SUB, X: ast.NewIdent("x")}),
		).
		Build()

	result := formatBlockStmt(t, body)
	expected := "{\n\tif x > 0 {\n\t\treturn x\n\t} else {\n\t\treturn -x\n\t}\n}"
	assert.Equal(t, expected, result)
}

func TestBodyBuilder_Call(t *testing.T) {
	body := NewBodyBuilder().
		Call(
			&ast.SelectorExpr{X: ast.NewIdent("fmt"), Sel: ast.NewIdent("Println")},
			&ast.BasicLit{Kind: token.STRING, Value: `"hello"`},
		).
		Build()

	result := formatBlockStmt(t, body)
	assert.Equal(t, "{\n\tfmt.Println(\"hello\")\n}", result)
}

func TestBodyBuilder_UtilityMethods(t *testing.T) {
	bb := NewBodyBuilder()
	assert.True(t, bb.IsEmpty())
	assert.Equal(t, 0, bb.StatementCount())

	bb.Return()
	assert.False(t, bb.IsEmpty())
	assert.Equal(t, 1, bb.StatementCount())

	bb.Clear()
	assert.True(t, bb.IsEmpty())
}

func TestBodyBuilder_Clone(t *testing.T) {
	original := NewBodyBuilder().Return()
	clone := original.Clone()
	clone.Return1(ast.NewIdent("nil"))

	assert.Equal(t, 1, original.StatementCount())
	assert.Equal(t, 2, clone.StatementCount())
}

func TestFunctionBuilder_WithBody(t *testing.T) {
	body := NewBodyBuilder().
		DeclareVar("result", ast.NewIdent("string")).
		Return1(ast.NewIdent("result"))

	fn := NewFunctionBuilder().
		WithName("GetResult").
		AddResult(StringField("")).
		WithBody(body).
		Build()

	result := formatFuncDecl(t, fn)
	expected := "func GetResult() string {\n\tvar result string\n\treturn result\n}"
	assert.Equal(t, expected, result)
}

func TestFunctionBuilder_DirectBodyManipulation(t *testing.T) {
	fn := NewFunctionBuilder().
		WithName("Process").
		AddResult(ErrorField())

	fn.Body().
		DeclareVar("err", ast.NewIdent("error")).
		Return1(ast.NewIdent("err"))

	result := formatFuncDecl(t, fn.Build())
	expected := "func Process() error {\n\tvar err error\n\treturn err\n}"
	assert.Equal(t, expected, result)
}
