package generator

import (
	"go/ast"
	"go/token"
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
		Names: []*ast.Ident{ast.NewIdent(name)},
		Type:  fieldType,
	}
	if tags != "" {
		ret.Tag = &ast.BasicLit{Kind: token.STRING, Value: tags}
	}

	return ret
}

func Star(field ast.Expr) ast.Expr {
	return &ast.StarExpr{X: field}
}

func Sel(field ast.Expr, sel string) ast.Expr {
	return &ast.SelectorExpr{
		X:   field,
		Sel: ast.NewIdent(sel),
	}
}
