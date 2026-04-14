package astbuilder

import (
	"go/ast"
)

// FunctionBuilder provides a fluent interface for building Go function declarations
type FunctionBuilder struct {
	name       string
	receiver   *FieldBuilder
	params     []*FieldBuilder
	rawParams  []*ast.Field
	results    []*FieldBuilder
	rawResults []*ast.Field
	body       *BodyBuilder
}

// NewFunctionBuilder creates a new FunctionBuilder
func NewFunctionBuilder() *FunctionBuilder {
	return &FunctionBuilder{
		params:     make([]*FieldBuilder, 0),
		rawParams:  make([]*ast.Field, 0),
		results:    make([]*FieldBuilder, 0),
		rawResults: make([]*ast.Field, 0),
		body:       NewBodyBuilder(),
	}
}

// WithName sets the function name
// Returns the builder for method chaining
func (fb *FunctionBuilder) WithName(name string) *FunctionBuilder {
	fb.name = name
	return fb
}

// WithReceiver sets the method receiver using a FieldBuilder
// Returns the builder for method chaining
func (fb *FunctionBuilder) WithReceiver(receiver *FieldBuilder) *FunctionBuilder {
	fb.receiver = receiver
	return fb
}

// AddParam adds a parameter using a FieldBuilder
// Returns the builder for method chaining
func (fb *FunctionBuilder) AddParam(param *FieldBuilder) *FunctionBuilder {
	if param == nil {
		panic("param cannot be nil")
	}
	fb.params = append(fb.params, param)
	return fb
}

// AddParams adds multiple parameters
// Returns the builder for method chaining
func (fb *FunctionBuilder) AddParams(params ...*FieldBuilder) *FunctionBuilder {
	for _, param := range params {
		if param == nil {
			panic("param cannot be nil")
		}
		fb.params = append(fb.params, param)
	}
	return fb
}

// AddResult adds a return value using a FieldBuilder
// Returns the builder for method chaining
func (fb *FunctionBuilder) AddResult(result *FieldBuilder) *FunctionBuilder {
	if result == nil {
		panic("result cannot be nil")
	}
	fb.results = append(fb.results, result)
	return fb
}

// AddResults adds multiple return values
// Returns the builder for method chaining
func (fb *FunctionBuilder) AddResults(results ...*FieldBuilder) *FunctionBuilder {
	for _, result := range results {
		if result == nil {
			panic("result cannot be nil")
		}
		fb.results = append(fb.results, result)
	}
	return fb
}

// AddParamExpr adds a parameter using a TypeExpressionBuilder for the type.
// Returns the builder for method chaining
func (fb *FunctionBuilder) AddParamExpr(name string, typeExpr TypeExpressionBuilder) *FunctionBuilder {
	field := &ast.Field{Type: typeExpr.Build()}
	if name != "" {
		field.Names = []*ast.Ident{ast.NewIdent(name)}
	}
	fb.rawParams = append(fb.rawParams, field)
	return fb
}

// AddResultExpr adds a return value using a TypeExpressionBuilder.
// Returns the builder for method chaining
func (fb *FunctionBuilder) AddResultExpr(typeExpr TypeExpressionBuilder) *FunctionBuilder {
	fb.rawResults = append(fb.rawResults, &ast.Field{Type: typeExpr.Build()})
	return fb
}

// WithBody sets the body builder
// Returns the builder for method chaining
func (fb *FunctionBuilder) WithBody(body *BodyBuilder) *FunctionBuilder {
	if body == nil {
		fb.body = NewBodyBuilder()
		return fb
	}
	fb.body = body
	return fb
}

// Body returns the body builder for direct manipulation
// Returns the body builder for method chaining on it
func (fb *FunctionBuilder) Body() *BodyBuilder {
	return fb.body
}

// Build creates the ast.FuncDecl
func (fb *FunctionBuilder) Build() *ast.FuncDecl {
	if fb.name == "" {
		panic("function must have a name")
	}

	// Convert params
	params := make([]*ast.Field, len(fb.params))
	for i, param := range fb.params {
		params[i] = param.Build()
	}
	params = append(params, fb.rawParams...)

	// Convert results
	results := make([]*ast.Field, len(fb.results))
	for i, result := range fb.results {
		results[i] = result.Build()
	}
	results = append(results, fb.rawResults...)

	funcDecl := &ast.FuncDecl{
		Name: ast.NewIdent(fb.name),
		Type: &ast.FuncType{
			Params:  &ast.FieldList{List: params},
			Results: &ast.FieldList{List: results},
		},
		Body: fb.body.Build(),
	}

	// Add receiver if present
	if fb.receiver != nil {
		funcDecl.Recv = &ast.FieldList{List: []*ast.Field{fb.receiver.Build()}}
	}

	return funcDecl
}

// Utility methods

// BuildDecl implements DeclBuilder.
func (fb *FunctionBuilder) BuildDecl() ast.Decl { return fb.Build() }

// HasName returns true if the function has a name
func (fb *FunctionBuilder) HasName() bool {
	return fb.name != ""
}

// GetName returns the function name
func (fb *FunctionBuilder) GetName() string {
	return fb.name
}

// HasReceiver returns true if the function has a receiver (is a method)
func (fb *FunctionBuilder) HasReceiver() bool {
	return fb.receiver != nil
}

// ParamCount returns the number of parameters
func (fb *FunctionBuilder) ParamCount() int {
	return len(fb.params) + len(fb.rawParams)
}

// ResultCount returns the number of return values
func (fb *FunctionBuilder) ResultCount() int {
	return len(fb.results) + len(fb.rawResults)
}

// Helper methods for common parameter types

// AddContextParam adds a context.Context parameter
func (fb *FunctionBuilder) AddContextParam(name string) *FunctionBuilder {
	return fb.AddParam(ContextField(name))
}

// AddErrorResult adds an error return value
func (fb *FunctionBuilder) AddErrorResult() *FunctionBuilder {
	return fb.AddResult(ErrorField())
}

// WithPointerReceiver sets a pointer receiver with the given name and type
func (fb *FunctionBuilder) WithPointerReceiver(name, typeName string) *FunctionBuilder {
	return fb.WithReceiver(NewFieldBuilder().
		WithName(name).
		WithType(Ident(typeName).AsPointer(true)))
}

// WithValueReceiver sets a value receiver with the given name and type
func (fb *FunctionBuilder) WithValueReceiver(name, typeName string) *FunctionBuilder {
	return fb.WithReceiver(IdentField(name, typeName))
}

// BodyBuilder provides a fluent interface for building function body statements
type BodyBuilder struct {
	statements []ast.Stmt
}

// NewBodyBuilder creates a new BodyBuilder
func NewBodyBuilder() *BodyBuilder {
	return &BodyBuilder{
		statements: make([]ast.Stmt, 0),
	}
}

// AddStmt adds a statement using a StatementBuilder
// Returns the builder for method chaining
func (bb *BodyBuilder) AddStmt(sb StatementBuilder) *BodyBuilder {
	if sb == nil {
		panic("statement builder cannot be nil")
	}
	bb.statements = append(bb.statements, sb.Build())
	return bb
}

// AddStmts adds multiple statements using StatementBuilders
// Returns the builder for method chaining
func (bb *BodyBuilder) AddStmts(sbs ...StatementBuilder) *BodyBuilder {
	for _, sb := range sbs {
		if sb == nil {
			panic("statement builder cannot be nil")
		}
		bb.statements = append(bb.statements, sb.Build())
	}
	return bb
}

// AddStatement adds a statement using a StatementBuilder.
// Returns the builder for method chaining
func (bb *BodyBuilder) AddStatement(sb StatementBuilder) *BodyBuilder {
	if sb == nil {
		panic("statement cannot be nil")
	}
	bb.statements = append(bb.statements, sb.Build())
	return bb
}

// AddStatements adds multiple statements using StatementBuilders.
// Returns the builder for method chaining
func (bb *BodyBuilder) AddStatements(sbs ...StatementBuilder) *BodyBuilder {
	for _, sb := range sbs {
		if sb == nil {
			panic("statement cannot be nil")
		}
		bb.statements = append(bb.statements, sb.Build())
	}
	return bb
}

// Return adds a return statement with no values
func (bb *BodyBuilder) Return() *BodyBuilder {
	return bb.AddStmt(NewReturnBuilder())
}

// Return1 adds a return statement with one value
func (bb *BodyBuilder) Return1(expr TypeExpressionBuilder) *BodyBuilder {
	return bb.AddStmt(Return1(expr))
}

// Return2 adds a return statement with two values
func (bb *BodyBuilder) Return2(expr1, expr2 TypeExpressionBuilder) *BodyBuilder {
	return bb.AddStmt(Return2(expr1, expr2))
}

// ReturnN adds a return statement with multiple values
func (bb *BodyBuilder) ReturnN(exprs ...TypeExpressionBuilder) *BodyBuilder {
	rb := NewReturnBuilder()
	for _, expr := range exprs {
		rb.AddResult(expr)
	}
	return bb.AddStmt(rb)
}

// Assign adds an assignment statement (=)
func (bb *BodyBuilder) Assign(lhs, rhs TypeExpressionBuilder) *BodyBuilder {
	return bb.AddStmt(Assign(lhs, rhs))
}

// Define adds a short variable declaration (:=)
func (bb *BodyBuilder) Define(lhs, rhs TypeExpressionBuilder) *BodyBuilder {
	return bb.AddStmt(Define(lhs, rhs))
}

// DeclareVar adds a variable declaration statement
func (bb *BodyBuilder) DeclareVar(name string, tb TypeExpressionBuilder) *BodyBuilder {
	return bb.AddStmt(DeclareVar(name, tb))
}

// If adds an if statement
func (bb *BodyBuilder) If(cond TypeExpressionBuilder, body *BodyBuilder) *BodyBuilder {
	return bb.AddStmt(NewIfBuilder().WithCond(cond).WithBody(body))
}

// IfElse adds an if-else statement
func (bb *BodyBuilder) IfElse(cond TypeExpressionBuilder, body *BodyBuilder, elseBody *BodyBuilder) *BodyBuilder {
	return bb.AddStmt(NewIfBuilder().WithCond(cond).WithBody(body).WithElse(elseBody))
}

// ExprStmt adds an expression statement (typically a function call)
func (bb *BodyBuilder) ExprStmt(expr TypeExpressionBuilder) *BodyBuilder {
	return bb.AddStmt(NewExprStmtBuilder().WithExpr(expr))
}

// Call adds a function/method call as a statement
func (bb *BodyBuilder) Call(fun TypeExpressionBuilder, args ...TypeExpressionBuilder) *BodyBuilder {
	return bb.AddStmt(CallStmt(fun, args...))
}

// Build creates the ast.BlockStmt
func (bb *BodyBuilder) Build() *ast.BlockStmt {
	return &ast.BlockStmt{List: bb.statements}
}

// Utility methods

// StatementCount returns the number of statements
func (bb *BodyBuilder) StatementCount() int {
	return len(bb.statements)
}

// IsEmpty returns true if the body has no statements
func (bb *BodyBuilder) IsEmpty() bool {
	return len(bb.statements) == 0
}

// Clear removes all statements
func (bb *BodyBuilder) Clear() *BodyBuilder {
	bb.statements = make([]ast.Stmt, 0)
	return bb
}

// Helper function builders

// Function creates a function builder with the given name
func Function(name string) *FunctionBuilder {
	return NewFunctionBuilder().WithName(name)
}

// Method creates a method builder with pointer receiver
func Method(receiverName, receiverType, methodName string) *FunctionBuilder {
	return NewFunctionBuilder().
		WithName(methodName).
		WithPointerReceiver(receiverName, receiverType)
}
