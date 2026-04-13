package generator

import (
	"go/ast"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-faster/errors"

	"github.com/sintoniastrategy/validgo-gen/internal/generator/astbuilder"
)

func (g *Generator) AddCreateResponseModel(baseName string, code string, response *openapi3.ResponseRef) error {
	fnBuilder := astbuilder.NewFunctionBuilder().
		WithName(baseName + code + "Response").
		AddResultExpr(astbuilder.Star(astbuilder.Selector(g.GetCurrentModelsPackage(), baseName+"Response")))

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
			astType = astbuilder.Sel(astbuilder.I(g.GetCurrentModelsPackage()),   typeName).Build()
			if json.Schema.Ref != "" {
				var importPath string
				typeName, importPath = g.ParseRefTypeName(json.Schema.Ref)
				if refIsExternal(json.Schema.Ref) {
					astType = astbuilder.I(typeName).Build()
				} else {
					astType = astbuilder.Sel(astbuilder.I(g.GetCurrentModelsPackage()),   typeName).Build()
				}
				if importPath != "" {
					g.AddHandlersImport(importPath)
				}
			}
			fnBuilder.AddParamExpr("body", astType)
			constructorArgs = append(constructorArgs, astbuilder.KeyValue(astbuilder.I("Body").Build(), astbuilder.I("body").Build()))
		}
	}

	if len(response.Value.Headers) > 0 {
		fnBuilder.AddParamExpr("headers", astbuilder.Sel(astbuilder.I(g.GetCurrentModelsPackage()),   baseName+"Response"+code+"Headers").Build())
		constructorArgs = append(constructorArgs, astbuilder.KeyValue(astbuilder.I("Headers").Build(), astbuilder.I("headers").Build()))
	}

	bodyBuilder := astbuilder.NewBodyBuilder().
		AddStmt(astbuilder.Return1(astbuilder.Amp(astbuilder.CompositeLit(
			astbuilder.Selector(g.GetCurrentModelsPackage(), baseName+"Response"),
			astbuilder.KeyValue(astbuilder.I("StatusCode").Build(), astbuilder.IntLit(code).Build()),
			astbuilder.KeyValue(astbuilder.I("Response"+code).Build(), astbuilder.Amp(astbuilder.CompositeLit(
				astbuilder.Selector(g.GetCurrentModelsPackage(), baseName+"Response"+code),
				constructorArgs...,
			)).Build()),
		)).Build()))

	fnBuilder.WithBody(bodyBuilder)
	g.HandlersFile.restBuilders = append(g.HandlersFile.restBuilders, fnBuilder)

	return nil
}
