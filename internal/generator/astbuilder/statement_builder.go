package astbuilder

import (
	"go/ast"
	"go/token"
)

// StatementBuilder is an interface for building ast.Stmt
type StatementBuilder interface {
	Build() ast.Stmt
}

// ReturnBuilder builds return statements
type ReturnBuilder struct {
	results []ast.Expr
}

// NewReturnBuilder creates a new ReturnBuilder
func NewReturnBuilder() *ReturnBuilder {
	return &ReturnBuilder{
		results: make([]ast.Expr, 0),
	}
}

// AddResult adds a return value expression
func (rb *ReturnBuilder) AddResult(expr ast.Expr) *ReturnBuilder {
	if expr == nil {
		panic("result expression cannot be nil")
	}
	rb.results = append(rb.results, expr)
	return rb
}

// AddResults adds multiple return value expressions
func (rb *ReturnBuilder) AddResults(exprs ...ast.Expr) *ReturnBuilder {
	for _, expr := range exprs {
		if expr == nil {
			panic("result expression cannot be nil")
		}
		rb.results = append(rb.results, expr)
	}
	return rb
}

// Build creates the ast.ReturnStmt
func (rb *ReturnBuilder) Build() ast.Stmt {
	return &ast.ReturnStmt{Results: rb.results}
}

// Clone creates a copy of the ReturnBuilder
func (rb *ReturnBuilder) Clone() *ReturnBuilder {
	clone := &ReturnBuilder{
		results: make([]ast.Expr, len(rb.results)),
	}
	copy(clone.results, rb.results)
	return clone
}

// Helper functions for ReturnBuilder

// Return creates an empty return statement
func Return() *ReturnBuilder {
	return NewReturnBuilder()
}

// Return1 creates a return statement with one value
func Return1(expr ast.Expr) *ReturnBuilder {
	return NewReturnBuilder().AddResult(expr)
}

// Return2 creates a return statement with two values
func Return2(expr1, expr2 ast.Expr) *ReturnBuilder {
	return NewReturnBuilder().AddResults(expr1, expr2)
}

// ReturnNil creates a return statement returning nil
func ReturnNil() *ReturnBuilder {
	return Return1(ast.NewIdent("nil"))
}

// ReturnErr creates a return statement returning err
func ReturnErr() *ReturnBuilder {
	return Return1(ast.NewIdent("err"))
}

// ReturnNilErr creates a return statement returning nil, err
func ReturnNilErr() *ReturnBuilder {
	return Return2(ast.NewIdent("nil"), ast.NewIdent("err"))
}

// AssignBuilder builds assignment statements
type AssignBuilder struct {
	lhs []ast.Expr
	rhs []ast.Expr
	tok token.Token
}

// NewAssignBuilder creates a new AssignBuilder with = token
func NewAssignBuilder() *AssignBuilder {
	return &AssignBuilder{
		lhs: make([]ast.Expr, 0),
		rhs: make([]ast.Expr, 0),
		tok: token.ASSIGN,
	}
}

// NewDefineBuilder creates a new AssignBuilder with := token
func NewDefineBuilder() *AssignBuilder {
	return &AssignBuilder{
		lhs: make([]ast.Expr, 0),
		rhs: make([]ast.Expr, 0),
		tok: token.DEFINE,
	}
}

// AddLhs adds a left-hand side expression
func (ab *AssignBuilder) AddLhs(expr ast.Expr) *AssignBuilder {
	if expr == nil {
		panic("lhs expression cannot be nil")
	}
	ab.lhs = append(ab.lhs, expr)
	return ab
}

// AddRhs adds a right-hand side expression
func (ab *AssignBuilder) AddRhs(expr ast.Expr) *AssignBuilder {
	if expr == nil {
		panic("rhs expression cannot be nil")
	}
	ab.rhs = append(ab.rhs, expr)
	return ab
}

// Lhs sets the left-hand side expressions (replaces existing)
func (ab *AssignBuilder) Lhs(exprs ...ast.Expr) *AssignBuilder {
	ab.lhs = make([]ast.Expr, 0, len(exprs))
	for _, expr := range exprs {
		if expr == nil {
			panic("lhs expression cannot be nil")
		}
		ab.lhs = append(ab.lhs, expr)
	}
	return ab
}

// Rhs sets the right-hand side expressions (replaces existing)
func (ab *AssignBuilder) Rhs(exprs ...ast.Expr) *AssignBuilder {
	ab.rhs = make([]ast.Expr, 0, len(exprs))
	for _, expr := range exprs {
		if expr == nil {
			panic("rhs expression cannot be nil")
		}
		ab.rhs = append(ab.rhs, expr)
	}
	return ab
}

// Build creates the ast.AssignStmt
func (ab *AssignBuilder) Build() ast.Stmt {
	if len(ab.lhs) == 0 {
		panic("assignment must have at least one lhs expression")
	}
	if len(ab.rhs) == 0 {
		panic("assignment must have at least one rhs expression")
	}
	return &ast.AssignStmt{
		Lhs: ab.lhs,
		Tok: ab.tok,
		Rhs: ab.rhs,
	}
}

// Clone creates a copy of the AssignBuilder
func (ab *AssignBuilder) Clone() *AssignBuilder {
	clone := &AssignBuilder{
		lhs: make([]ast.Expr, len(ab.lhs)),
		rhs: make([]ast.Expr, len(ab.rhs)),
		tok: ab.tok,
	}
	copy(clone.lhs, ab.lhs)
	copy(clone.rhs, ab.rhs)
	return clone
}

// Helper functions for AssignBuilder

// Assign creates an assignment statement: lhs = rhs
func Assign(lhs, rhs ast.Expr) *AssignBuilder {
	return NewAssignBuilder().AddLhs(lhs).AddRhs(rhs)
}

// Define creates a short variable declaration: lhs := rhs
func Define(lhs, rhs ast.Expr) *AssignBuilder {
	return NewDefineBuilder().AddLhs(lhs).AddRhs(rhs)
}

// Define2 creates a short variable declaration with two lhs: lhs1, lhs2 := rhs
func Define2(lhs1, lhs2, rhs ast.Expr) *AssignBuilder {
	return NewDefineBuilder().Lhs(lhs1, lhs2).AddRhs(rhs)
}

// DefineCall creates: result := funcCall(args...)
func DefineCall(result string, fun ast.Expr, args ...ast.Expr) *AssignBuilder {
	return NewDefineBuilder().
		AddLhs(ast.NewIdent(result)).
		AddRhs(Call(fun, args...))
}

// DefineCallWithErr creates: result, err := funcCall(args...)
func DefineCallWithErr(result string, fun ast.Expr, args ...ast.Expr) *AssignBuilder {
	return NewDefineBuilder().
		Lhs(ast.NewIdent(result), ast.NewIdent("err")).
		AddRhs(Call(fun, args...))
}

// VarDeclBuilder builds variable declaration statements
type VarDeclBuilder struct {
	name     string
	typeExpr ast.Expr
	value    ast.Expr
}

// NewVarDeclBuilder creates a new VarDeclBuilder
func NewVarDeclBuilder() *VarDeclBuilder {
	return &VarDeclBuilder{}
}

// WithName sets the variable name
func (vdb *VarDeclBuilder) WithName(name string) *VarDeclBuilder {
	vdb.name = name
	return vdb
}

// WithType sets the variable type
func (vdb *VarDeclBuilder) WithType(typeExpr ast.Expr) *VarDeclBuilder {
	vdb.typeExpr = typeExpr
	return vdb
}

// WithTypeBuilder sets the variable type using a TypeExpressionBuilder
func (vdb *VarDeclBuilder) WithTypeBuilder(tb TypeExpressionBuilder) *VarDeclBuilder {
	vdb.typeExpr = tb.Build()
	return vdb
}

// WithValue sets the initial value
func (vdb *VarDeclBuilder) WithValue(value ast.Expr) *VarDeclBuilder {
	vdb.value = value
	return vdb
}

// Build creates the ast.DeclStmt
func (vdb *VarDeclBuilder) Build() ast.Stmt {
	if vdb.name == "" {
		panic("variable declaration must have a name")
	}
	if vdb.typeExpr == nil && vdb.value == nil {
		panic("variable declaration must have a type or value")
	}

	spec := &ast.ValueSpec{
		Names: []*ast.Ident{ast.NewIdent(vdb.name)},
	}
	if vdb.typeExpr != nil {
		spec.Type = vdb.typeExpr
	}
	if vdb.value != nil {
		spec.Values = []ast.Expr{vdb.value}
	}

	return &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok:   token.VAR,
			Specs: []ast.Spec{spec},
		},
	}
}

// Clone creates a copy of the VarDeclBuilder
func (vdb *VarDeclBuilder) Clone() *VarDeclBuilder {
	return &VarDeclBuilder{
		name:     vdb.name,
		typeExpr: vdb.typeExpr,
		value:    vdb.value,
	}
}

// Helper functions for VarDeclBuilder

// DeclareVar creates a variable declaration: var name Type
func DeclareVar(name string, typeExpr ast.Expr) *VarDeclBuilder {
	return NewVarDeclBuilder().WithName(name).WithType(typeExpr)
}

// DeclareVarWithValue creates a variable declaration with value: var name Type = value
func DeclareVarWithValue(name string, typeExpr, value ast.Expr) *VarDeclBuilder {
	return NewVarDeclBuilder().WithName(name).WithType(typeExpr).WithValue(value)
}

// DeclareVarWithType creates a variable declaration using a TypeExpressionBuilder: var name Type
func DeclareVarWithType(name string, tb TypeExpressionBuilder) *VarDeclBuilder {
	return NewVarDeclBuilder().WithName(name).WithTypeBuilder(tb)
}

// IfBuilder builds if statements
type IfBuilder struct {
	init     ast.Stmt
	cond     ast.Expr
	body     *BodyBuilder
	elseBody *BodyBuilder
	elseIf   *IfBuilder
}

// NewIfBuilder creates a new IfBuilder
func NewIfBuilder() *IfBuilder {
	return &IfBuilder{
		body: NewBodyBuilder(),
	}
}

// WithInit sets the init statement (e.g., if x := foo(); x != nil)
func (ib *IfBuilder) WithInit(init ast.Stmt) *IfBuilder {
	ib.init = init
	return ib
}

// WithInitBuilder sets the init statement using a StatementBuilder
func (ib *IfBuilder) WithInitBuilder(sb StatementBuilder) *IfBuilder {
	ib.init = sb.Build()
	return ib
}

// WithCond sets the condition expression
func (ib *IfBuilder) WithCond(cond ast.Expr) *IfBuilder {
	ib.cond = cond
	return ib
}

// WithBody sets the if body
func (ib *IfBuilder) WithBody(body *BodyBuilder) *IfBuilder {
	if body == nil {
		ib.body = NewBodyBuilder()
		return ib
	}
	ib.body = body.Clone()
	return ib
}

// Body returns the body builder for direct manipulation
func (ib *IfBuilder) Body() *BodyBuilder {
	return ib.body
}

// WithElse sets the else body
func (ib *IfBuilder) WithElse(elseBody *BodyBuilder) *IfBuilder {
	if elseBody == nil {
		ib.elseBody = nil
		return ib
	}
	ib.elseBody = elseBody.Clone()
	return ib
}

// Else returns the else body builder for direct manipulation
func (ib *IfBuilder) Else() *BodyBuilder {
	if ib.elseBody == nil {
		ib.elseBody = NewBodyBuilder()
	}
	return ib.elseBody
}

// WithElseIf sets an else-if clause
func (ib *IfBuilder) WithElseIf(elseIf *IfBuilder) *IfBuilder {
	ib.elseIf = elseIf
	return ib
}

// Build creates the ast.IfStmt
func (ib *IfBuilder) Build() ast.Stmt {
	if ib.cond == nil {
		panic("if statement must have a condition")
	}

	stmt := &ast.IfStmt{
		Init: ib.init,
		Cond: ib.cond,
		Body: ib.body.Build(),
	}

	if ib.elseIf != nil {
		stmt.Else = ib.elseIf.Build()
	} else if ib.elseBody != nil {
		stmt.Else = ib.elseBody.Build()
	}

	return stmt
}

// Clone creates a copy of the IfBuilder
func (ib *IfBuilder) Clone() *IfBuilder {
	clone := &IfBuilder{
		init: ib.init,
		cond: ib.cond,
	}
	if ib.body != nil {
		clone.body = ib.body.Clone()
	}
	if ib.elseBody != nil {
		clone.elseBody = ib.elseBody.Clone()
	}
	if ib.elseIf != nil {
		clone.elseIf = ib.elseIf.Clone()
	}
	return clone
}

// Helper functions for IfBuilder

// If creates an if statement
func If(cond ast.Expr) *IfBuilder {
	return NewIfBuilder().WithCond(cond)
}

// IfErrNotNil creates: if err != nil { ... }
func IfErrNotNil() *IfBuilder {
	return If(&ast.BinaryExpr{
		X:  ast.NewIdent("err"),
		Op: token.NEQ,
		Y:  ast.NewIdent("nil"),
	})
}

// IfErrNotNilReturn creates: if err != nil { return ..., err }
func IfErrNotNilReturn(returnValues ...ast.Expr) *IfBuilder {
	rb := NewReturnBuilder()
	for _, v := range returnValues {
		rb.AddResult(v)
	}
	rb.AddResult(ast.NewIdent("err"))
	return IfErrNotNil().WithBody(NewBodyBuilder().AddStmt(rb))
}

// IfNil creates: if expr == nil { ... }
func IfNil(expr ast.Expr) *IfBuilder {
	return If(&ast.BinaryExpr{
		X:  expr,
		Op: token.EQL,
		Y:  ast.NewIdent("nil"),
	})
}

// IfNotNil creates: if expr != nil { ... }
func IfNotNil(expr ast.Expr) *IfBuilder {
	return If(&ast.BinaryExpr{
		X:  expr,
		Op: token.NEQ,
		Y:  ast.NewIdent("nil"),
	})
}

// ExprStmtBuilder builds expression statements
type ExprStmtBuilder struct {
	expr ast.Expr
}

// NewExprStmtBuilder creates a new ExprStmtBuilder
func NewExprStmtBuilder() *ExprStmtBuilder {
	return &ExprStmtBuilder{}
}

// WithExpr sets the expression
func (esb *ExprStmtBuilder) WithExpr(expr ast.Expr) *ExprStmtBuilder {
	esb.expr = expr
	return esb
}

// Build creates the ast.ExprStmt
func (esb *ExprStmtBuilder) Build() ast.Stmt {
	if esb.expr == nil {
		panic("expression statement must have an expression")
	}
	return &ast.ExprStmt{X: esb.expr}
}

// Clone creates a copy of the ExprStmtBuilder
func (esb *ExprStmtBuilder) Clone() *ExprStmtBuilder {
	return &ExprStmtBuilder{expr: esb.expr}
}

// Helper functions for ExprStmtBuilder

// ExprStmt creates an expression statement
func ExprStmt(expr ast.Expr) *ExprStmtBuilder {
	return NewExprStmtBuilder().WithExpr(expr)
}

// Call creates a function call expression: fun(args...)
// Returns *ast.CallExpr which implements ast.Expr, for use as a sub-expression.
func Call(fun ast.Expr, args ...ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{Fun: fun, Args: args}
}

// KeyValue creates a key-value expression: key: value
// Used in composite literals for struct and map fields.
func KeyValue(key, value ast.Expr) *ast.KeyValueExpr {
	return &ast.KeyValueExpr{Key: key, Value: value}
}

// CompositeLit creates a composite literal: Type{elts...}
// The type is built from a TypeExpressionBuilder (Ident, Selector, MapOf, etc.).
func CompositeLit(tb TypeExpressionBuilder, elts ...ast.Expr) *ast.CompositeLit {
	return &ast.CompositeLit{Type: tb.Build(), Elts: elts}
}

// Index creates an index expression: x[index]
func Index(x, index ast.Expr) *ast.IndexExpr {
	return &ast.IndexExpr{X: x, Index: index}
}

// Or creates a logical OR expression: x || y
func Or(x, y ast.Expr) *ast.BinaryExpr {
	return &ast.BinaryExpr{X: x, Op: token.LOR, Y: y}
}

// And creates a logical AND expression: x && y
func And(x, y ast.Expr) *ast.BinaryExpr {
	return &ast.BinaryExpr{X: x, Op: token.LAND, Y: y}
}

// Add creates an addition expression: x + y
func Add(x, y ast.Expr) *ast.BinaryExpr {
	return &ast.BinaryExpr{X: x, Op: token.ADD, Y: y}
}

// Not creates a logical NOT expression: !x
func Not(x ast.Expr) *ast.UnaryExpr {
	return &ast.UnaryExpr{Op: token.NOT, X: x}
}

// CallStmt creates a function call statement
func CallStmt(fun ast.Expr, args ...ast.Expr) *ExprStmtBuilder {
	return ExprStmt(Call(fun, args...))
}

// MethodCallStmt creates a method call statement: receiver.method(args...)
func MethodCallStmt(receiver, method string, args ...ast.Expr) *ExprStmtBuilder {
	return CallStmt(
		&ast.SelectorExpr{
			X:   ast.NewIdent(receiver),
			Sel: ast.NewIdent(method),
		},
		args...,
	)
}

// RangeBuilder builds range statements
type RangeBuilder struct {
	key      ast.Expr
	value    ast.Expr
	x        ast.Expr
	body     *BodyBuilder
	isDefine bool
}

// NewRangeBuilder creates a new RangeBuilder
func NewRangeBuilder() *RangeBuilder {
	return &RangeBuilder{
		body:     NewBodyBuilder(),
		isDefine: true,
	}
}

// WithKey sets the key variable
func (rb *RangeBuilder) WithKey(key ast.Expr) *RangeBuilder {
	rb.key = key
	return rb
}

// WithValue sets the value variable
func (rb *RangeBuilder) WithValue(value ast.Expr) *RangeBuilder {
	rb.value = value
	return rb
}

// Over sets the expression to range over
func (rb *RangeBuilder) Over(x ast.Expr) *RangeBuilder {
	rb.x = x
	return rb
}

// WithBody sets the loop body
func (rb *RangeBuilder) WithBody(body *BodyBuilder) *RangeBuilder {
	if body == nil {
		rb.body = NewBodyBuilder()
		return rb
	}
	rb.body = body.Clone()
	return rb
}

// Body returns the body builder for direct manipulation
func (rb *RangeBuilder) Body() *BodyBuilder {
	return rb.body
}

// AsAssign uses = instead of := for the loop variables
func (rb *RangeBuilder) AsAssign() *RangeBuilder {
	rb.isDefine = false
	return rb
}

// Build creates the ast.RangeStmt
func (rb *RangeBuilder) Build() ast.Stmt {
	if rb.x == nil {
		panic("range statement must have an expression to range over")
	}

	tok := token.DEFINE
	if !rb.isDefine {
		tok = token.ASSIGN
	}

	return &ast.RangeStmt{
		Key:   rb.key,
		Value: rb.value,
		Tok:   tok,
		X:     rb.x,
		Body:  rb.body.Build(),
	}
}

// Clone creates a copy of the RangeBuilder
func (rb *RangeBuilder) Clone() *RangeBuilder {
	clone := &RangeBuilder{
		key:      rb.key,
		value:    rb.value,
		x:        rb.x,
		isDefine: rb.isDefine,
	}
	if rb.body != nil {
		clone.body = rb.body.Clone()
	}
	return clone
}

// Helper functions for RangeBuilder

// Range creates a range statement: for key, value := range x
func Range(key, value string, x ast.Expr) *RangeBuilder {
	rb := NewRangeBuilder().Over(x)
	if key != "" && key != "_" {
		rb.WithKey(ast.NewIdent(key))
	} else if key == "_" {
		rb.WithKey(ast.NewIdent("_"))
	}
	if value != "" && value != "_" {
		rb.WithValue(ast.NewIdent(value))
	} else if value == "_" {
		rb.WithValue(ast.NewIdent("_"))
	}
	return rb
}

// RangeValue creates a range statement: for _, value := range x
func RangeValue(value string, x ast.Expr) *RangeBuilder {
	return Range("_", value, x)
}

// RangeIndex creates a range statement: for i := range x
func RangeIndex(index string, x ast.Expr) *RangeBuilder {
	return Range(index, "", x)
}

// SwitchBuilder builds switch statements
type SwitchBuilder struct {
	init  ast.Stmt
	tag   ast.Expr
	cases []*CaseBuilder
}

// NewSwitchBuilder creates a new SwitchBuilder
func NewSwitchBuilder() *SwitchBuilder {
	return &SwitchBuilder{
		cases: make([]*CaseBuilder, 0),
	}
}

// WithInit sets the init statement
func (sb *SwitchBuilder) WithInit(init ast.Stmt) *SwitchBuilder {
	sb.init = init
	return sb
}

// WithTag sets the switch tag expression
func (sb *SwitchBuilder) WithTag(tag ast.Expr) *SwitchBuilder {
	sb.tag = tag
	return sb
}

// AddCase adds a case clause
func (sb *SwitchBuilder) AddCase(cb *CaseBuilder) *SwitchBuilder {
	if cb == nil {
		panic("case builder cannot be nil")
	}
	sb.cases = append(sb.cases, cb)
	return sb
}

// Build creates the ast.SwitchStmt
func (sb *SwitchBuilder) Build() ast.Stmt {
	clauses := make([]ast.Stmt, len(sb.cases))
	for i, c := range sb.cases {
		clauses[i] = c.Build()
	}

	return &ast.SwitchStmt{
		Init: sb.init,
		Tag:  sb.tag,
		Body: &ast.BlockStmt{List: clauses},
	}
}

// Clone creates a copy of the SwitchBuilder
func (sb *SwitchBuilder) Clone() *SwitchBuilder {
	clone := &SwitchBuilder{
		init:  sb.init,
		tag:   sb.tag,
		cases: make([]*CaseBuilder, len(sb.cases)),
	}
	for i, c := range sb.cases {
		clone.cases[i] = c.Clone()
	}
	return clone
}

// CaseBuilder builds case clauses
type CaseBuilder struct {
	exprs []ast.Expr
	body  *BodyBuilder
}

// NewCaseBuilder creates a new CaseBuilder
func NewCaseBuilder() *CaseBuilder {
	return &CaseBuilder{
		exprs: make([]ast.Expr, 0),
		body:  NewBodyBuilder(),
	}
}

// AddExpr adds a case expression
func (cb *CaseBuilder) AddExpr(expr ast.Expr) *CaseBuilder {
	if expr == nil {
		panic("case expression cannot be nil")
	}
	cb.exprs = append(cb.exprs, expr)
	return cb
}

// WithBody sets the case body
func (cb *CaseBuilder) WithBody(body *BodyBuilder) *CaseBuilder {
	if body == nil {
		cb.body = NewBodyBuilder()
		return cb
	}
	cb.body = body.Clone()
	return cb
}

// Body returns the body builder for direct manipulation
func (cb *CaseBuilder) Body() *BodyBuilder {
	return cb.body
}

// Build creates the ast.CaseClause (returns ast.Stmt for interface)
func (cb *CaseBuilder) Build() ast.Stmt {
	var list []ast.Expr
	if len(cb.exprs) > 0 {
		list = cb.exprs
	}
	return &ast.CaseClause{
		List: list,
		Body: cb.body.Build().List,
	}
}

// Clone creates a copy of the CaseBuilder
func (cb *CaseBuilder) Clone() *CaseBuilder {
	clone := &CaseBuilder{
		exprs: make([]ast.Expr, len(cb.exprs)),
	}
	copy(clone.exprs, cb.exprs)
	if cb.body != nil {
		clone.body = cb.body.Clone()
	}
	return clone
}

// Helper functions for SwitchBuilder

// Switch creates a switch statement with a tag
func Switch(tag ast.Expr) *SwitchBuilder {
	return NewSwitchBuilder().WithTag(tag)
}

// Case creates a case clause with expressions
func Case(exprs ...ast.Expr) *CaseBuilder {
	cb := NewCaseBuilder()
	for _, expr := range exprs {
		cb.AddExpr(expr)
	}
	return cb
}

// Default creates a default case clause
func Default() *CaseBuilder {
	return NewCaseBuilder()
}
