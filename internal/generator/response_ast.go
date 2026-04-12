package generator

import (
	"go/ast"
	"go/token"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-faster/errors"

	"github.com/sintoniastrategy/validgo-gen/internal/generator/astbuilder"
)

func (g *Generator) AddCreateResponseModel(baseName string, code string, response *openapi3.ResponseRef) error {
	fnBuilder := astbuilder.NewFunctionBuilder().
		WithName(baseName + code + "Response").
		AddResult(astbuilder.NewFieldBuilder().WithType(astbuilder.Selector(g.GetCurrentModelsPackage(), baseName+"Response").AsPointer(true)))

	constructorArgs := []ast.Expr{}

	if len(response.Value.Content) > 0 {
		// assume there is a json body
		json, ok := response.Value.Content["application/json"]
		if !ok {
			return errors.New("response content type 'application/json' not found")
		}
		if json.Schema != nil {
			typeName := baseName + "Response" + code + "Body"
			var astType ast.Expr
			astType = Sel(I(g.GetCurrentModelsPackage()), typeName)
			if json.Schema.Ref != "" {
				var importPath string
				typeName, importPath = g.ParseRefTypeName(json.Schema.Ref)
				if refIsExternal(json.Schema.Ref) {
					astType = I(typeName)
				} else {
					astType = Sel(I(g.GetCurrentModelsPackage()), typeName)
				}
				if importPath != "" {
					g.AddHandlersImport(importPath)
				}
			}
			// Add param manually since we have a dynamic type
			fn := fnBuilder.Build()
			fn.Type.Params.List = append(fn.Type.Params.List, &ast.Field{
				Names: []*ast.Ident{I("body")},
				Type:  astType,
			})
			fnBuilder = astbuilder.NewFunctionBuilder().WithName(baseName + code + "Response")
			fnBuilder.Build().Type = fn.Type
			fnBuilder.Build().Recv = fn.Recv

			constructorArgs = append(constructorArgs, &ast.KeyValueExpr{Key: I("Body"), Value: I("body")})
		}
	}

	if len(response.Value.Headers) > 0 {
		constructorArgs = append(constructorArgs, &ast.KeyValueExpr{Key: I("Headers"), Value: I("headers")})
	}

	bodyBuilder := astbuilder.NewBodyBuilder().
		AddStmt(astbuilder.Return1(Amp(&ast.CompositeLit{
			Type: Sel(I(g.GetCurrentModelsPackage()), baseName+"Response"),
			Elts: []ast.Expr{
				&ast.KeyValueExpr{
					Key:   I("StatusCode"),
					Value: &ast.BasicLit{Kind: token.INT, Value: code},
				},
				&ast.KeyValueExpr{
					Key: I("Response" + code),
					Value: Amp(&ast.CompositeLit{
						Type: Sel(I(g.GetCurrentModelsPackage()), baseName+"Response"+code),
						Elts: constructorArgs,
					}),
				},
			},
		})))

	// Build manually since params have dynamic types
	arglist := []*ast.Field{}
	if len(response.Value.Content) > 0 {
		json, ok := response.Value.Content["application/json"]
		if ok && json.Schema != nil {
			typeName := baseName + "Response" + code + "Body"
			var astType ast.Expr
			astType = Sel(I(g.GetCurrentModelsPackage()), typeName)
			if json.Schema.Ref != "" {
				var importPath string
				typeName, importPath = g.ParseRefTypeName(json.Schema.Ref)
				if refIsExternal(json.Schema.Ref) {
					astType = I(typeName)
				} else {
					astType = Sel(I(g.GetCurrentModelsPackage()), typeName)
				}
				if importPath != "" {
					g.AddHandlersImport(importPath)
				}
			}
			arglist = append(arglist, &ast.Field{Names: []*ast.Ident{I("body")}, Type: astType})
		}
	}
	if len(response.Value.Headers) > 0 {
		arglist = append(arglist, &ast.Field{
			Names: []*ast.Ident{I("headers")},
			Type:  Sel(I(g.GetCurrentModelsPackage()), baseName+"Response"+code+"Headers"),
		})
	}

	fn := &ast.FuncDecl{
		Name: I(baseName + code + "Response"),
		Type: &ast.FuncType{
			Params:  &ast.FieldList{List: arglist},
			Results: &ast.FieldList{List: []*ast.Field{{Type: Star(Sel(I(g.GetCurrentModelsPackage()), baseName+"Response"))}}},
		},
		Body: bodyBuilder.Build(),
	}

	g.HandlersFile.restDecls = append(g.HandlersFile.restDecls, fn)

	return nil
}
