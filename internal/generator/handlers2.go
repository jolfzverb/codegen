package generator

import (
	"fmt"
	"go/ast"
	"go/token"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-faster/errors"

	"github.com/jolfzverb/codegen/internal/generator/astbuilder"
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
		WithName("parse" + baseName + "QueryParams").
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
		WithName("parse" + baseName + "Headers").
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
		WithName("parse" + baseName + "Cookies").
		WithPointerReceiver("h", "Handler").
		AddParam(astbuilder.NewFieldBuilder().WithName("r").WithType(astbuilder.Selector("http", "Request").AsPointer(true))).
		AddResult(astbuilder.NewFieldBuilder().WithType(astbuilder.Selector(g.GetCurrentModelsPackage(), baseName+"Cookies").AsPointer(true))).
		AddResult(astbuilder.ErrorField()).
		WithBody(bodyBuilder).
		Build()

	g.HandlersFile.restDecls = append(g.HandlersFile.restDecls, fn)

	return nil
}

func (g *Generator) GetValidateFuncStmt(typeName string, ref string) ast.Expr {
	validateFuncName := "Validate" + typeName + "JSON"

	if ref == "" || !refIsExternal(ref) {
		return I(validateFuncName)
	}

	filename := parseFilenameFromRef(ref)
	if filename == "" {
		return I(validateFuncName)
	}

	parts := strings.Split(ref, "/")
	if len(parts) == 0 {
		return I(validateFuncName)
	}

	validateFuncName = "Validate" + parts[len(parts)-1] + "JSON"

	g.YAMLFilesToProcess = append(g.YAMLFilesToProcess, g.GetYAMLFilePath(filename))
	g.AddHandlersImport(g.GetHandlersImportForFile(filename))
	modelName := g.GetModelName(filename)
	return Sel(I(modelName), validateFuncName)
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

	fn := astbuilder.Function("parse" + baseName + "RequestBody").
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
		WithName("parse" + baseName + "Request").
		WithPointerReceiver("h", "Handler").
		AddParam(astbuilder.NewFieldBuilder().WithName("r").WithType(astbuilder.Selector("http", "Request").AsPointer(true))).
		AddResult(astbuilder.NewFieldBuilder().WithType(astbuilder.Selector(g.GetCurrentModelsPackage(), baseName+"Request").AsPointer(true))).
		AddResult(astbuilder.ErrorField()).
		WithBody(bodyBuilder).
		Build()

	g.HandlersFile.restDecls = append(g.HandlersFile.restDecls, fn)
}

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

func (g *Generator) AddContainsNullIfNeeded() {
	if g.HandlersFile.hasContainsNullMethod {
		return
	}

	g.HandlersFile.hasContainsNullMethod = true

	bodyBuilder := astbuilder.NewBodyBuilder().
		AddStatement(&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{I("temp")},
						Type:  I("any"),
					},
				},
			},
		}).
		AddStmt(astbuilder.DefineCall("err", Sel(I("json"), "Unmarshal"), I("data"), Amp(I("temp")))).
		AddStmt(astbuilder.If(Ne(I("err"), I("nil"))).WithBody(astbuilder.NewBodyBuilder().
			AddStmt(astbuilder.Return1(I("false"))))).
		AddStmt(astbuilder.Return1(&ast.BinaryExpr{X: I("temp"), Op: token.EQL, Y: I("nil")}))

	fn := astbuilder.Function("containsNull").
		AddParam(astbuilder.SelectorField("data", "json", "RawMessage")).
		AddResult(astbuilder.BoolField("")).
		WithBody(bodyBuilder).
		Build()

	g.HandlersFile.restDecls = append(g.HandlersFile.restDecls, fn)
	g.AddHandlersImport("encoding/json")
}

func (g *Generator) AddObjectValidate(modelName string, schema *openapi3.SchemaRef) error {
	const op = "generator.AddObjectValidate"
	requiredFieldsMap := make(map[string]bool, 0)
	nullableFields := make([]string, 0)
	objectFields := make(map[string]ast.Expr, 0)

	for _, requiredField := range schema.Value.Required {
		requiredFieldsMap[requiredField] = true
	}

	for fieldName, fieldSchema := range schema.Value.Properties {
		if fieldSchema.Value == nil {
			continue
		}
		if fieldSchema.Value.Nullable && requiredFieldsMap[fieldName] {
			nullableFields = append(nullableFields, fieldName)
		}
		if fieldSchema.Value.Type.Permits(openapi3.TypeObject) {
			fieldType, err := g.GetFieldTypeFromSchema(modelName, fieldName, fieldSchema)
			if err != nil {
				return errors.Wrap(err, op)
			}
			objectFields[fieldName] = g.GetValidateFuncStmt(fieldType, fieldSchema.Ref)
		}
		if fieldSchema.Value.Type.Permits(openapi3.TypeArray) {
			if fieldSchema.Value.Items != nil {
				itemsType := g.getMostNestedArrayItemType(fieldSchema.Value.Items)
				if itemsType != nil && itemsType.Permits(openapi3.TypeObject) {
					fieldType, err := g.GetFieldTypeFromSchema(modelName, fieldName, fieldSchema)
					if err != nil {
						return errors.Wrap(err, op)
					}
					objectFields[fieldName] = g.GetValidateFuncStmt(fieldType, fieldSchema.Ref)
				}
			}
		}
	}
	requiredFields := make([]string, 0, len(requiredFieldsMap))
	for fieldName := range requiredFieldsMap {
		requiredFields = append(requiredFields, fieldName)
	}
	sort.Strings(requiredFields)
	sort.Strings(nullableFields)

	bodyBuilder := astbuilder.NewBodyBuilder()

	if len(requiredFields) > 0 {
		requiredFieldsElts := make([]ast.Expr, 0, len(requiredFields))
		for _, fieldName := range requiredFields {
			requiredFieldsElts = append(requiredFieldsElts, &ast.KeyValueExpr{Key: Str(fieldName), Value: I("true")})
		}
		bodyBuilder.AddStmt(astbuilder.Define(I("requiredFields"), &ast.CompositeLit{
			Type: &ast.MapType{Key: I("string"), Value: I("bool")},
			Elts: requiredFieldsElts,
		}))

		nullableFieldsElts := make([]ast.Expr, 0, len(nullableFields))
		for _, fieldName := range nullableFields {
			nullableFieldsElts = append(nullableFieldsElts, &ast.KeyValueExpr{Key: Str(fieldName), Value: I("true")})
		}
		bodyBuilder.AddStmt(astbuilder.Define(I("nullableFields"), &ast.CompositeLit{
			Type: &ast.MapType{Key: I("string"), Value: I("bool")},
			Elts: nullableFieldsElts,
		}))
	}

	if len(requiredFields) > 0 || len(objectFields) > 0 {
		bodyBuilder.AddStatement(&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{I("obj")},
						Type:  &ast.MapType{Key: I("string"), Value: Sel(I("json"), "RawMessage")},
					},
				},
			},
		})
		bodyBuilder.
			AddStmt(astbuilder.DefineCall("err", Sel(I("json"), "Unmarshal"), I("jsonData"), Amp(I("obj")))).
			AddStmt(astbuilder.IfErrNotNil().WithBody(astbuilder.NewBodyBuilder().AddStmt(astbuilder.Return1(I("err")))))
	}

	if len(requiredFields) > 0 || len(objectFields) > 0 {
		bodyBuilder.AddStatement(&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{&ast.ValueSpec{Names: []*ast.Ident{I("val")}, Type: Sel(I("json"), "RawMessage")}},
			},
		})
		bodyBuilder.AddStatement(&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{&ast.ValueSpec{Names: []*ast.Ident{I("exists")}, Type: I("bool")}},
			},
		})
	}

	if len(requiredFields) > 0 {
		rangeBody := astbuilder.NewBodyBuilder().
			AddStmt(astbuilder.NewAssignBuilder().Lhs(I("val"), I("exists")).Rhs(&ast.IndexExpr{X: I("obj"), Index: I("field")})).
			AddStmt(astbuilder.If(&ast.UnaryExpr{Op: token.NOT, X: I("exists")}).WithBody(astbuilder.NewBodyBuilder().
				AddStmt(astbuilder.Return1(&ast.CallExpr{
					Fun: Sel(I("errors"), "New"),
					Args: []ast.Expr{&ast.BinaryExpr{
						X:  &ast.BinaryExpr{X: Str("field "), Op: token.ADD, Y: I("field")},
						Op: token.ADD,
						Y:  Str(" is required"),
					}},
				})))).
			AddStmt(astbuilder.If(&ast.BinaryExpr{
				X:  &ast.UnaryExpr{Op: token.NOT, X: &ast.IndexExpr{X: I("nullableFields"), Index: I("field")}},
				Op: token.LAND,
				Y:  &ast.CallExpr{Fun: I("containsNull"), Args: []ast.Expr{I("val")}},
			}).WithBody(astbuilder.NewBodyBuilder().
				AddStmt(astbuilder.Return1(&ast.CallExpr{
					Fun: Sel(I("errors"), "New"),
					Args: []ast.Expr{&ast.BinaryExpr{
						X:  &ast.BinaryExpr{X: Str("field "), Op: token.ADD, Y: I("field")},
						Op: token.ADD,
						Y:  Str(" cannot be null"),
					}},
				}))))

		bodyBuilder.AddStmt(astbuilder.RangeIndex("field", I("requiredFields")).WithBody(rangeBody))
		g.AddContainsNullIfNeeded()
		g.AddHandlersImport("github.com/go-faster/errors")
	}

	objectFieldsNames := make([]string, 0, len(objectFields))
	for fieldName := range objectFields {
		objectFieldsNames = append(objectFieldsNames, fieldName)
	}
	sort.Strings(objectFieldsNames)

	for _, fieldName := range objectFieldsNames {
		fieldValidationFunc := objectFields[fieldName]
		bodyBuilder.AddStmt(astbuilder.NewAssignBuilder().Lhs(I("val"), I("exists")).Rhs(&ast.IndexExpr{X: I("obj"), Index: Str(fieldName)}))

		ifBody := astbuilder.NewBodyBuilder().
			AddStmt(astbuilder.Assign(I("err"), &ast.CallExpr{Fun: fieldValidationFunc, Args: []ast.Expr{I("val")}})).
			AddStmt(astbuilder.IfErrNotNil().WithBody(astbuilder.NewBodyBuilder().
				AddStmt(astbuilder.Return1(&ast.CallExpr{
					Fun:  Sel(I("errors"), "Wrap"),
					Args: []ast.Expr{I("err"), Str("field " + fieldName + " is not valid")},
				}))))

		bodyBuilder.AddStmt(astbuilder.If(&ast.BinaryExpr{
			X:  I("exists"),
			Op: token.LAND,
			Y:  &ast.UnaryExpr{Op: token.NOT, X: &ast.CallExpr{Fun: I("containsNull"), Args: []ast.Expr{I("val")}}},
		}).WithBody(ifBody))

		g.AddContainsNullIfNeeded()
		g.AddHandlersImport("github.com/go-faster/errors")
	}

	bodyBuilder.AddStmt(astbuilder.Return1(I("nil")))

	paramName := "jsonData"
	if len(requiredFields) == 0 && len(objectFields) == 0 {
		paramName = "_"
	}

	fn := astbuilder.Function("Validate" + modelName + "JSON").
		AddParam(astbuilder.SelectorField(paramName, "json", "RawMessage")).
		AddResult(astbuilder.ErrorField()).
		WithBody(bodyBuilder).
		Build()

	g.HandlersFile.restDecls = append(g.HandlersFile.restDecls, fn)
	return nil
}

func (g *Generator) AddArrayValidate(modelName string, schema *openapi3.SchemaRef) error {
	const op = "generator.AddArrayValidate"

	elemType, err := g.GetFieldTypeFromSchema(modelName, "Item", schema.Value.Items)
	if err != nil {
		return errors.Wrap(err, op)
	}
	validateFunc := g.GetValidateFuncStmt(elemType, schema.Value.Items.Ref)

	rangeIfBody := astbuilder.NewBodyBuilder().
		AddStmt(astbuilder.Assign(I("err"), &ast.CallExpr{Fun: validateFunc, Args: []ast.Expr{I("obj")}})).
		AddStmt(astbuilder.IfErrNotNil().WithBody(astbuilder.NewBodyBuilder().
			AddStmt(astbuilder.Return1(&ast.CallExpr{
				Fun:  Sel(I("errors"), "Wrapf"),
				Args: []ast.Expr{I("err"), Str("error validating object at index %d"), I("index")},
			}))))

	rangeBody := astbuilder.NewBodyBuilder().
		AddStmt(astbuilder.If(&ast.UnaryExpr{
			Op: token.NOT,
			X:  &ast.CallExpr{Fun: I("containsNull"), Args: []ast.Expr{I("obj")}},
		}).WithBody(rangeIfBody))

	bodyBuilder := astbuilder.NewBodyBuilder().
		AddStatement(&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{I("arr")},
						Type:  &ast.ArrayType{Elt: Sel(I("json"), "RawMessage")},
					},
				},
			},
		}).
		AddStmt(astbuilder.DefineCall("err", Sel(I("json"), "Unmarshal"), I("jsonData"), Amp(I("arr")))).
		AddStmt(astbuilder.IfErrNotNil().WithBody(astbuilder.NewBodyBuilder().AddStmt(astbuilder.Return1(I("err"))))).
		AddStmt(astbuilder.Range("index", "obj", I("arr")).WithBody(rangeBody)).
		AddStmt(astbuilder.Return1(I("nil")))

	fn := astbuilder.Function("Validate" + modelName + "JSON").
		AddParam(astbuilder.SelectorField("jsonData", "json", "RawMessage")).
		AddResult(astbuilder.ErrorField()).
		WithBody(bodyBuilder).
		Build()

	g.HandlersFile.restDecls = append(g.HandlersFile.restDecls, fn)
	g.AddHandlersImport("github.com/go-faster/errors")

	return nil
}
