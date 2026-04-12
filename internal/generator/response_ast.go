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
		AddResultExpr(Star(Sel(I(g.GetCurrentModelsPackage()), baseName+"Response")))

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
			fnBuilder.AddParamExpr("body", astType)
			constructorArgs = append(constructorArgs, astbuilder.KeyValue(I("Body"), I("body")))
		}
	}

	if len(response.Value.Headers) > 0 {
		fnBuilder.AddParamExpr("headers", Sel(I(g.GetCurrentModelsPackage()), baseName+"Response"+code+"Headers"))
		constructorArgs = append(constructorArgs, astbuilder.KeyValue(I("Headers"), I("headers")))
	}

	bodyBuilder := astbuilder.NewBodyBuilder().
		AddStmt(astbuilder.Return1(Amp(&ast.CompositeLit{
			Type: Sel(I(g.GetCurrentModelsPackage()), baseName+"Response"),
			Elts: []ast.Expr{
				astbuilder.KeyValue(I("StatusCode"), &ast.BasicLit{Kind: token.INT, Value: code}),
				astbuilder.KeyValue(I("Response"+code), Amp(&ast.CompositeLit{
					Type: Sel(I(g.GetCurrentModelsPackage()), baseName+"Response"+code),
					Elts: constructorArgs,
				})),
			},
		})))

	g.HandlersFile.restDecls = append(g.HandlersFile.restDecls, fnBuilder.WithBody(bodyBuilder).Build())

	return nil
}
