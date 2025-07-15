package generator

import (
	"go/ast"
	"go/token"
	"strconv"
)

func Func(name string, receiver *ast.Field, params []*ast.Field, result []*ast.Field, body []ast.Stmt) *ast.FuncDecl {
	ret := &ast.FuncDecl{
		Name: ast.NewIdent(name),
		Type: &ast.FuncType{
			Params:  &ast.FieldList{List: params},
			Results: &ast.FieldList{List: result},
		},
		Body: &ast.BlockStmt{List: body},
	}
	if receiver != nil {
		ret.Recv = &ast.FieldList{List: []*ast.Field{receiver}}
	}

	return ret
}

func Field(name string, fieldType ast.Expr, tags string) *ast.Field {
	ret := &ast.Field{
		Type: fieldType,
	}
	if name != "" {
		ret.Names = []*ast.Ident{ast.NewIdent(name)}
	}
	if tags != "" {
		ret.Tag = &ast.BasicLit{Kind: token.STRING, Value: tags}
	}

	return ret
}

func FieldA(field *ast.Field) []*ast.Field {
	return []*ast.Field{field}
}

func Star(field ast.Expr) *ast.StarExpr {
	return &ast.StarExpr{X: field}
}

func Sel(field ast.Expr, sel string) *ast.SelectorExpr {
	return &ast.SelectorExpr{
		X:   field,
		Sel: ast.NewIdent(sel),
	}
}

func Amp(field ast.Expr) *ast.UnaryExpr {
	return &ast.UnaryExpr{
		Op: token.AND,
		X:  field,
	}
}

func I(name string) *ast.Ident {
	return ast.NewIdent(name)
}

func Str(value string) *ast.BasicLit {
	return &ast.BasicLit{
		Kind:  token.STRING,
		Value: strconv.Quote(value),
	}
}

func Ne(left, right ast.Expr) *ast.BinaryExpr {
	return &ast.BinaryExpr{
		X:  left,
		Op: token.NEQ,
		Y:  right,
	}
}

func Eq(left, right ast.Expr) *ast.BinaryExpr {
	return &ast.BinaryExpr{
		X:  left,
		Op: token.EQL,
		Y:  right,
	}
}

func Ret() *ast.ReturnStmt {
	return &ast.ReturnStmt{
		Results: []ast.Expr{},
	}
}

func Ret1(expr ast.Expr) *ast.ReturnStmt {
	return &ast.ReturnStmt{
		Results: []ast.Expr{expr},
	}
}

func Ret2(expr1, expr2 ast.Expr) *ast.ReturnStmt {
	return &ast.ReturnStmt{
		Results: []ast.Expr{expr1, expr2},
	}
}
