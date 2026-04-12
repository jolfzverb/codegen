package generator

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-faster/errors"

	"github.com/sintoniastrategy/validgo-gen/internal/generator/astbuilder"
)

func (g *Generator) AddParseQueryParamsMethod(baseName string, params openapi3.Parameters) error {
	bodyBuilder := astbuilder.NewBodyBuilder().
		AddStatement(&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{I("queryParams")},
						Type:  Sel(I(g.GetCurrentModelsPackage()), baseName+"QueryParams"),
					},
				},
			},
		})

	for _, param := range params {
		if param.Value.Schema == nil || param.Value.Schema.Value == nil {
			continue
		}

		varName := GoIdentLowercase(FormatGoLikeIdentifier(param.Value.Name))
		bodyBuilder.AddStmt(astbuilder.Define(I(varName), &ast.CallExpr{
			Fun: Sel(&ast.CallExpr{
				Fun:  Sel(Sel(I("r"), "URL"), "Query"),
				Args: []ast.Expr{},
			}, "Get"),
			Args: []ast.Expr{Str(param.Value.Name)},
		}))

		if param.Value.Required {
			bodyBuilder.AddStmt(astbuilder.If(Eq(I(varName), Str(""))).WithBody(astbuilder.NewBodyBuilder().
				AddStmt(astbuilder.Return2(I("nil"), &ast.CallExpr{
					Fun:  Sel(I("errors"), "New"),
					Args: []ast.Expr{Str(param.Value.Name + " query param is required")},
				}))))
			g.AddHandlersImport("github.com/go-faster/errors")
			switch {
			case param.Value.Schema.Value.Type.Permits("string"):
				for _, stmt := range g.AssignStringField("queryParams", varName, FormatGoLikeIdentifier(param.Value.Name), param.Value.Schema, param.Value.Required) {
					bodyBuilder.AddStatement(stmt)
				}
			default:
				return errors.New(fmt.Sprintf("unsupported path parameter type: %v", param.Value.Schema.Value.Type)) //nolint:revive
			}
		} else {
			ifBody := astbuilder.NewBodyBuilder()
			for _, stmt := range g.AssignStringField("queryParams", varName, FormatGoLikeIdentifier(param.Value.Name), param.Value.Schema, param.Value.Required) {
				ifBody.AddStatement(stmt)
			}
			bodyBuilder.AddStmt(astbuilder.If(Ne(I(varName), Str(""))).WithBody(ifBody))
		}
	}

	bodyBuilder.
		AddStmt(astbuilder.DefineCall("err", Sel(Sel(I("h"), "validator"), "Struct"), I("queryParams"))).
		AddStmt(astbuilder.IfErrNotNilReturn(I("nil"))).
		AddStmt(astbuilder.Return2(Amp(I("queryParams")), I("nil")))

	fn := astbuilder.NewFunctionBuilder().
		WithName("parse"+baseName+"QueryParams").
		WithPointerReceiver("h", "Handler").
		AddParam(astbuilder.NewFieldBuilder().WithName("r").WithType(astbuilder.Selector("http", "Request").AsPointer(true))).
		AddResult(astbuilder.NewFieldBuilder().WithType(astbuilder.Selector(g.GetCurrentModelsPackage(), baseName+"QueryParams").AsPointer(true))).
		AddResult(astbuilder.ErrorField()).
		WithBody(bodyBuilder).
		Build()

	g.HandlersFile.restDecls = append(g.HandlersFile.restDecls, fn)

	return nil
}

func (g *Generator) AssignStringField(paramsName string, varName string, fieldName string, param *openapi3.SchemaRef, required bool) []ast.Stmt {
	if param.Value.Format == "date-time" {
		g.AddHandlersImport("time")
		bodyBuilder := astbuilder.NewBodyBuilder().
			AddStmt(astbuilder.DefineCallWithErr("parsed"+fieldName, Sel(I("time"), "Parse"), Sel(I("time"), "RFC3339"), I(varName))).
			AddStmt(astbuilder.IfErrNotNil().WithBody(astbuilder.NewBodyBuilder().
				AddStmt(astbuilder.Return2(I("nil"), &ast.CallExpr{
					Fun:  Sel(I("errors"), "Wrap"),
					Args: []ast.Expr{I("err"), Str(fieldName + " is not a valid date-time format")},
				}))))

		var rhs ast.Expr
		if required && !g.HandlersFile.requiredFieldsArePointers {
			rhs = I("parsed" + fieldName)
		} else {
			rhs = Amp(I("parsed" + fieldName))
		}
		bodyBuilder.AddStmt(astbuilder.Assign(Sel(I(paramsName), fieldName), rhs))

		return bodyBuilder.Build().List
	}

	var rhs ast.Expr
	if required && !g.HandlersFile.requiredFieldsArePointers {
		rhs = I(varName)
	} else {
		rhs = Amp(I(varName))
	}

	return []ast.Stmt{astbuilder.Assign(Sel(I(paramsName), fieldName), rhs).Build()}
}

func (g *Generator) AddParseHeadersMethod(baseName string, params openapi3.Parameters) error {
	bodyBuilder := astbuilder.NewBodyBuilder().
		AddStatement(&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{I("headers")},
						Type:  Sel(I(g.GetCurrentModelsPackage()), baseName+"Headers"),
					},
				},
			},
		})

	for _, param := range params {
		if param.Value.Schema == nil || param.Value.Schema.Value == nil {
			continue
		}
		if g.Opts.AllowRemoteAddrParam && param.Value.Name == "Remote-Addr" && param.Value.Schema.Value.Format == "remote-addr" {
			bodyBuilder.AddStmt(astbuilder.Assign(
				Sel(I("headers"), FormatGoLikeIdentifier(param.Value.Name)),
				Sel(I("r"), "RemoteAddr"),
			))
			continue
		}
		varName := GoIdentLowercase(FormatGoLikeIdentifier(param.Value.Name))
		bodyBuilder.AddStmt(astbuilder.Define(I(varName), &ast.CallExpr{
			Fun:  Sel(Sel(I("r"), "Header"), "Get"),
			Args: []ast.Expr{Str(param.Value.Name)},
		}))

		if param.Value.Required {
			bodyBuilder.AddStmt(astbuilder.If(Eq(I(varName), Str(""))).WithBody(astbuilder.NewBodyBuilder().
				AddStmt(astbuilder.Return2(I("nil"), &ast.CallExpr{
					Fun:  Sel(I("errors"), "New"),
					Args: []ast.Expr{Str(param.Value.Name + " header is required")},
				}))))
			g.AddHandlersImport("github.com/go-faster/errors")
			switch {
			case param.Value.Schema.Value.Type.Permits("string"):
				for _, stmt := range g.AssignStringField("headers", varName, FormatGoLikeIdentifier(param.Value.Name), param.Value.Schema, param.Value.Required) {
					bodyBuilder.AddStatement(stmt)
				}
			default:
				return errors.New("unsupported path parameter type: " + fmt.Sprint(param.Value.Schema.Value.Type))
			}
		} else {
			ifBody := astbuilder.NewBodyBuilder()
			for _, stmt := range g.AssignStringField("headers", varName, FormatGoLikeIdentifier(param.Value.Name), param.Value.Schema, param.Value.Required) {
				ifBody.AddStatement(stmt)
			}
			bodyBuilder.AddStmt(astbuilder.If(Ne(I(varName), Str(""))).WithBody(ifBody))
		}
	}

	bodyBuilder.
		AddStmt(astbuilder.DefineCall("err", Sel(Sel(I("h"), "validator"), "Struct"), I("headers"))).
		AddStmt(astbuilder.IfErrNotNilReturn(I("nil"))).
		AddStmt(astbuilder.Return2(Amp(I("headers")), I("nil")))

	fn := astbuilder.NewFunctionBuilder().
		WithName("parse"+baseName+"Headers").
		WithPointerReceiver("h", "Handler").
		AddParam(astbuilder.NewFieldBuilder().WithName("r").WithType(astbuilder.Selector("http", "Request").AsPointer(true))).
		AddResult(astbuilder.NewFieldBuilder().WithType(astbuilder.Selector(g.GetCurrentModelsPackage(), baseName+"Headers").AsPointer(true))).
		AddResult(astbuilder.ErrorField()).
		WithBody(bodyBuilder).
		Build()

	g.HandlersFile.restDecls = append(g.HandlersFile.restDecls, fn)

	return nil
}

func (g *Generator) AddParseCookiesMethod(baseName string, params openapi3.Parameters) error {
	bodyBuilder := astbuilder.NewBodyBuilder().
		AddStatement(&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{I("cookies")},
						Type:  Sel(I(g.GetCurrentModelsPackage()), baseName+"Cookies"),
					},
				},
			},
		})

	for _, param := range params {
		if param.Value.Schema == nil || param.Value.Schema.Value == nil {
			continue
		}

		varName := GoIdentLowercase(FormatGoLikeIdentifier(param.Value.Name))
		bodyBuilder.AddStmt(astbuilder.DefineCallWithErr(varName, Sel(I("r"), "Cookie"), Str(param.Value.Name)))

		if param.Value.Required {
			bodyBuilder.AddStmt(astbuilder.IfErrNotNilReturn(I("nil")))
		} else {
			bodyBuilder.AddStmt(astbuilder.If(&ast.BinaryExpr{
				X:  Ne(I("err"), I("nil")),
				Op: token.LAND,
				Y: &ast.UnaryExpr{
					Op: token.NOT,
					X: &ast.CallExpr{
						Fun:  Sel(I("errors"), "Is"),
						Args: []ast.Expr{I("err"), Sel(I("http"), "ErrNoCookie")},
					},
				},
			}).WithBody(astbuilder.NewBodyBuilder().AddStmt(astbuilder.Return2(I("nil"), I("err")))))
			g.AddHandlersImport("github.com/go-faster/errors")
		}

		if param.Value.Required {
			bodyBuilder.AddStmt(astbuilder.Define(I(varName+"Value"), Sel(I(varName), "Value")))

			switch {
			case param.Value.Schema.Value.Type.Permits("string"):
				for _, stmt := range g.AssignStringField("cookies", varName+"Value", FormatGoLikeIdentifier(param.Value.Name), param.Value.Schema, param.Value.Required) {
					bodyBuilder.AddStatement(stmt)
				}
			default:
				return errors.New("unsupported path parameter type: " + fmt.Sprint(param.Value.Schema.Value.Type))
			}
		} else {
			ifBody := astbuilder.NewBodyBuilder().
				AddStmt(astbuilder.Define(I(varName+"Value"), Sel(I(varName), "Value")))
			for _, stmt := range g.AssignStringField("cookies", varName+"Value", FormatGoLikeIdentifier(param.Value.Name), param.Value.Schema, param.Value.Required) {
				ifBody.AddStatement(stmt)
			}
			bodyBuilder.AddStmt(astbuilder.If(Eq(I("err"), I("nil"))).WithBody(ifBody))
		}
	}

	bodyBuilder.
		AddStmt(astbuilder.Assign(I("err"), &ast.CallExpr{
			Fun:  Sel(Sel(I("h"), "validator"), "Struct"),
			Args: []ast.Expr{I("cookies")},
		})).
		AddStmt(astbuilder.IfErrNotNilReturn(I("nil"))).
		AddStmt(astbuilder.Return2(Amp(I("cookies")), I("nil")))

	fn := astbuilder.NewFunctionBuilder().
		WithName("parse"+baseName+"Cookies").
		WithPointerReceiver("h", "Handler").
		AddParam(astbuilder.NewFieldBuilder().WithName("r").WithType(astbuilder.Selector("http", "Request").AsPointer(true))).
		AddResult(astbuilder.NewFieldBuilder().WithType(astbuilder.Selector(g.GetCurrentModelsPackage(), baseName+"Cookies").AsPointer(true))).
		AddResult(astbuilder.ErrorField()).
		WithBody(bodyBuilder).
		Build()

	g.HandlersFile.restDecls = append(g.HandlersFile.restDecls, fn)

	return nil
}

func (g *Generator) AddParseRequestBodyMethod(baseName string, contentType string, body *openapi3.RequestBodyRef) error {
	bodyBuilder := astbuilder.NewBodyBuilder()

	if !body.Value.Required {
		bodyBuilder.AddStmt(astbuilder.IfNil(Sel(I("r"), "Body")).WithBody(astbuilder.NewBodyBuilder().
			AddStmt(astbuilder.Return2(I("nil"), I("nil")))))
	}

	typeName := baseName + "RequestBody"
	var bodyType ast.Expr
	content, ok := body.Value.Content[contentType]
	bodyType = Sel(I(g.GetCurrentModelsPackage()), typeName)
	if ok && content.Schema != nil {
		if content.Schema.Ref != "" {
			var importPath string
			typeName, importPath = g.ParseRefTypeName(content.Schema.Ref)
			bodyType = Sel(I(g.GetCurrentModelsPackage()), typeName)
			if importPath != "" {
				g.AddHandlersImport(importPath)
			}
			if refIsExternal(content.Schema.Ref) {
				bodyType = I(typeName)
			}
		}
	}

	bodyBuilder.AddStatement(&ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{I("bodyJSON")},
					Type:  Sel(I("json"), "RawMessage"),
				},
			},
		},
	})
	g.AddHandlersImport("encoding/json")

	bodyBuilder.
		AddStmt(astbuilder.DefineCall("err", Sel(&ast.CallExpr{
			Fun:  Sel(I("json"), "NewDecoder"),
			Args: []ast.Expr{Sel(I("r"), "Body")},
		}, "Decode"), Amp(I("bodyJSON")))).
		AddStmt(astbuilder.IfErrNotNilReturn(I("nil"))).
		AddStmt(astbuilder.Assign(I("err"), &ast.CallExpr{
			Fun:  g.GetValidateFuncStmt(typeName, content.Schema.Ref),
			Args: []ast.Expr{I("bodyJSON")},
		})).
		AddStmt(astbuilder.IfErrNotNilReturn(I("nil")))

	bodyBuilder.AddStatement(&ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{I("body")},
					Type:  bodyType,
				},
			},
		},
	})

	bodyBuilder.
		AddStmt(astbuilder.Assign(I("err"), &ast.CallExpr{
			Fun:  Sel(I("json"), "Unmarshal"),
			Args: []ast.Expr{I("bodyJSON"), Amp(I("body"))},
		})).
		AddStmt(astbuilder.IfErrNotNilReturn(I("nil"))).
		AddStmt(astbuilder.Assign(I("err"), &ast.CallExpr{
			Fun:  Sel(Sel(I("h"), "validator"), "Struct"),
			Args: []ast.Expr{I("body")},
		})).
		AddStmt(astbuilder.IfErrNotNilReturn(I("nil"))).
		AddStmt(astbuilder.Return2(Amp(I("body")), I("nil")))

	fn := astbuilder.Function("parse"+baseName+"RequestBody").
		WithPointerReceiver("h", "Handler").
		AddParam(astbuilder.NewFieldBuilder().WithName("r").WithType(astbuilder.Selector("http", "Request").AsPointer(true))).
		AddResult(astbuilder.ErrorField()).
		WithBody(bodyBuilder).
		Build()

	// Set the first result type manually since it's a dynamic type
	fn.Type.Results.List = []*ast.Field{
		{Type: Star(bodyType)},
		{Type: I("error")},
	}

	g.HandlersFile.restDecls = append(g.HandlersFile.restDecls, fn)

	return nil
}

func (g *Generator) AddParseRequestMethod(baseName string, contentType string, pathParams openapi3.Parameters,
	queryParams openapi3.Parameters, headers openapi3.Parameters, cookieParams openapi3.Parameters,
	body *openapi3.RequestBodyRef,
) {
	bodyBuilder := astbuilder.NewBodyBuilder()
	elts := []ast.Expr{}

	if len(pathParams) > 0 {
		elts = append(elts, &ast.KeyValueExpr{Key: I("Path"), Value: Star(I("pathParams"))})
		bodyBuilder.
			AddStmt(astbuilder.DefineCallWithErr("pathParams", Sel(I("h"), "parse"+baseName+"PathParams"), I("r"))).
			AddStmt(astbuilder.IfErrNotNilReturn(I("nil")))
	}
	if len(queryParams) > 0 {
		elts = append(elts, &ast.KeyValueExpr{Key: I("Query"), Value: Star(I("queryParams"))})
		bodyBuilder.
			AddStmt(astbuilder.DefineCallWithErr("queryParams", Sel(I("h"), "parse"+baseName+"QueryParams"), I("r"))).
			AddStmt(astbuilder.IfErrNotNilReturn(I("nil")))
	}
	if len(headers) > 0 {
		elts = append(elts, &ast.KeyValueExpr{Key: I("Headers"), Value: Star(I("headers"))})
		bodyBuilder.
			AddStmt(astbuilder.DefineCallWithErr("headers", Sel(I("h"), "parse"+baseName+"Headers"), I("r"))).
			AddStmt(astbuilder.IfErrNotNilReturn(I("nil")))
	}
	if len(cookieParams) > 0 {
		elts = append(elts, &ast.KeyValueExpr{Key: I("Cookies"), Value: Star(I("cookieParams"))})
		bodyBuilder.
			AddStmt(astbuilder.DefineCallWithErr("cookieParams", Sel(I("h"), "parse"+baseName+"Cookies"), I("r"))).
			AddStmt(astbuilder.IfErrNotNilReturn(I("nil")))
	}
	if body != nil && body.Value != nil {
		content, ok := body.Value.Content[contentType]
		if ok && content.Schema != nil {
			if body.Value.Required {
				elts = append(elts, &ast.KeyValueExpr{Key: I("Body"), Value: Star(I("body"))})
			} else {
				elts = append(elts, &ast.KeyValueExpr{Key: I("Body"), Value: I("body")})
			}
			bodyBuilder.
				AddStmt(astbuilder.DefineCallWithErr("body", Sel(I("h"), "parse"+baseName+"RequestBody"), I("r"))).
				AddStmt(astbuilder.IfErrNotNilReturn(I("nil")))
		}
	}

	bodyBuilder.AddStmt(astbuilder.Return2(
		Amp(&ast.CompositeLit{
			Type: Sel(I(g.GetCurrentModelsPackage()), baseName+"Request"),
			Elts: elts,
		}),
		I("nil"),
	))

	fn := astbuilder.NewFunctionBuilder().
		WithName("parse"+baseName+"Request").
		WithPointerReceiver("h", "Handler").
		AddParam(astbuilder.NewFieldBuilder().WithName("r").WithType(astbuilder.Selector("http", "Request").AsPointer(true))).
		AddResult(astbuilder.NewFieldBuilder().WithType(astbuilder.Selector(g.GetCurrentModelsPackage(), baseName+"Request").AsPointer(true))).
		AddResult(astbuilder.ErrorField()).
		WithBody(bodyBuilder).
		Build()

	g.HandlersFile.restDecls = append(g.HandlersFile.restDecls, fn)
}
