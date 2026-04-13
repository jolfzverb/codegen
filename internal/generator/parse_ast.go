package generator

import (
	"fmt"
	"go/ast"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-faster/errors"

	"github.com/sintoniastrategy/validgo-gen/internal/generator/astbuilder"
)

func (g *Generator) AddParseQueryParamsMethod(baseName string, params openapi3.Parameters) error {
	bodyBuilder := astbuilder.NewBodyBuilder().
		AddStmt(astbuilder.DeclareVar("queryParams", astbuilder.Sel(astbuilder.I(g.GetCurrentModelsPackage()),   baseName+"QueryParams").Build()))

	for _, param := range params {
		if param.Value.Schema == nil || param.Value.Schema.Value == nil {
			continue
		}

		varName := GoIdentLowercase(FormatGoLikeIdentifier(param.Value.Name))
		bodyBuilder.AddStmt(astbuilder.Define(astbuilder.I(varName).Build(), 
			astbuilder.Call(astbuilder.Sel(astbuilder.Call(astbuilder.Sel(astbuilder.Sel(astbuilder.I("r"),  "URL"),  "Query")),  "Get"), astbuilder.Str(param.Value.Name)).Build()))

		if param.Value.Required {
			bodyBuilder.AddStmt(astbuilder.If(astbuilder.Eq(astbuilder.I(varName), astbuilder.Str("")).Build()).WithBody(astbuilder.NewBodyBuilder().
				AddStmt(astbuilder.Return2(astbuilder.I("nil").Build(),  astbuilder.Call(astbuilder.Sel(astbuilder.I("errors"),  "New"), astbuilder.Str(param.Value.Name+" query param is required")).Build()))))
			g.AddHandlersImport("github.com/go-faster/errors")
			switch {
			case param.Value.Schema.Value.Type.Permits("string"):
				g.AssignStringField(bodyBuilder, "queryParams", varName, FormatGoLikeIdentifier(param.Value.Name), param.Value.Schema, param.Value.Required)
			default:
				return errors.New(fmt.Sprintf("unsupported path parameter type: %v", param.Value.Schema.Value.Type)) //nolint:revive
			}
		} else {
			ifBody := astbuilder.NewBodyBuilder()
			g.AssignStringField(ifBody, "queryParams", varName, FormatGoLikeIdentifier(param.Value.Name), param.Value.Schema, param.Value.Required)
			bodyBuilder.AddStmt(astbuilder.If(astbuilder.Ne(astbuilder.I(varName), astbuilder.Str("")).Build()).WithBody(ifBody))
		}
	}

	bodyBuilder.
		AddStmt(astbuilder.DefineCall("err", astbuilder.Sel(astbuilder.Sel(astbuilder.I("h"),   "validator"),   "Struct"), astbuilder.I("queryParams"))).
		AddStmt(astbuilder.IfErrNotNilReturn(astbuilder.I("nil").Build())).
		AddStmt(astbuilder.Return2(astbuilder.Amp(astbuilder.Ident("queryParams")).Build(),  astbuilder.I("nil").Build()))

	fn := astbuilder.NewFunctionBuilder().
		WithName("parse"+baseName+"QueryParams").
		WithPointerReceiver("h", "Handler").
		AddParam(astbuilder.NewFieldBuilder().WithName("r").WithType(astbuilder.Selector("http", "Request").AsPointer(true))).
		AddResult(astbuilder.NewFieldBuilder().WithType(astbuilder.Selector(g.GetCurrentModelsPackage(), baseName+"QueryParams").AsPointer(true))).
		AddResult(astbuilder.ErrorField()).
		WithBody(bodyBuilder)

	g.HandlersFile.restBuilders = append(g.HandlersFile.restBuilders, fn)

	return nil
}

func (g *Generator) AssignStringField(bodyBuilder *astbuilder.BodyBuilder, paramsName string, varName string, fieldName string, param *openapi3.SchemaRef, required bool) {
	if param.Value.Format == "date-time" {
		g.AddHandlersImport("time")
		bodyBuilder.
			AddStmt(astbuilder.DefineCallWithErr("parsed"+fieldName, astbuilder.Sel(astbuilder.I("time"),   "Parse"), astbuilder.Sel(astbuilder.I("time"),   "RFC3339"), astbuilder.I(varName))).
			AddStmt(astbuilder.IfErrNotNil().WithBody(astbuilder.NewBodyBuilder().
				AddStmt(astbuilder.Return2(astbuilder.I("nil").Build(),  astbuilder.Call(astbuilder.Sel(astbuilder.I("errors"),  "Wrap"), astbuilder.I("err"), astbuilder.Str(fieldName+" is not a valid date-time format")).Build()))))

		var rhs astbuilder.TypeExpressionBuilder
		if required && !g.HandlersFile.requiredFieldsArePointers {
			rhs = astbuilder.Ident("parsed" + fieldName)
		} else {
			rhs = astbuilder.Amp(astbuilder.Ident("parsed" + fieldName))
		}
		bodyBuilder.AddStmt(astbuilder.Assign(astbuilder.Sel(astbuilder.I(paramsName),  fieldName).Build(),  rhs.Build()))
		return
	}

	var rhs astbuilder.TypeExpressionBuilder
	if required && !g.HandlersFile.requiredFieldsArePointers {
		rhs = astbuilder.Ident(varName)
	} else {
		rhs = astbuilder.Amp(astbuilder.Ident(varName))
	}
	bodyBuilder.AddStmt(astbuilder.Assign(astbuilder.Sel(astbuilder.I(paramsName),  fieldName).Build(),  rhs.Build()))
}

func (g *Generator) AddParseHeadersMethod(baseName string, params openapi3.Parameters) error {
	bodyBuilder := astbuilder.NewBodyBuilder().
		AddStmt(astbuilder.DeclareVar("headers", astbuilder.Sel(astbuilder.I(g.GetCurrentModelsPackage()),   baseName+"Headers").Build()))

	for _, param := range params {
		if param.Value.Schema == nil || param.Value.Schema.Value == nil {
			continue
		}
		if g.Opts.AllowRemoteAddrParam && param.Value.Name == "Remote-Addr" && param.Value.Schema.Value.Format == "remote-addr" {
			bodyBuilder.AddStmt(astbuilder.Assign(
				astbuilder.Sel(astbuilder.I("headers"),  FormatGoLikeIdentifier(param.Value.Name)).Build(), 
				astbuilder.Sel(astbuilder.I("r"),  "RemoteAddr").Build()))
			continue
		}
		varName := GoIdentLowercase(FormatGoLikeIdentifier(param.Value.Name))
		bodyBuilder.AddStmt(astbuilder.Define(astbuilder.I(varName).Build(), 
			astbuilder.Call(astbuilder.Sel(astbuilder.Sel(astbuilder.I("r"),  "Header"),  "Get"), astbuilder.Str(param.Value.Name)).Build()))

		if param.Value.Required {
			bodyBuilder.AddStmt(astbuilder.If(astbuilder.Eq(astbuilder.I(varName), astbuilder.Str("")).Build()).WithBody(astbuilder.NewBodyBuilder().
				AddStmt(astbuilder.Return2(astbuilder.I("nil").Build(),  astbuilder.Call(astbuilder.Sel(astbuilder.I("errors"),  "New"), astbuilder.Str(param.Value.Name+" header is required")).Build()))))
			g.AddHandlersImport("github.com/go-faster/errors")
			switch {
			case param.Value.Schema.Value.Type.Permits("string"):
				g.AssignStringField(bodyBuilder, "headers", varName, FormatGoLikeIdentifier(param.Value.Name), param.Value.Schema, param.Value.Required)
			default:
				return errors.New("unsupported path parameter type: " + fmt.Sprint(param.Value.Schema.Value.Type))
			}
		} else {
			ifBody := astbuilder.NewBodyBuilder()
			g.AssignStringField(ifBody, "headers", varName, FormatGoLikeIdentifier(param.Value.Name), param.Value.Schema, param.Value.Required)
			bodyBuilder.AddStmt(astbuilder.If(astbuilder.Ne(astbuilder.I(varName), astbuilder.Str("")).Build()).WithBody(ifBody))
		}
	}

	bodyBuilder.
		AddStmt(astbuilder.DefineCall("err", astbuilder.Sel(astbuilder.Sel(astbuilder.I("h"),   "validator"),   "Struct"), astbuilder.I("headers"))).
		AddStmt(astbuilder.IfErrNotNilReturn(astbuilder.I("nil").Build())).
		AddStmt(astbuilder.Return2(astbuilder.Amp(astbuilder.Ident("headers")).Build(),  astbuilder.I("nil").Build()))

	fn := astbuilder.NewFunctionBuilder().
		WithName("parse"+baseName+"Headers").
		WithPointerReceiver("h", "Handler").
		AddParam(astbuilder.NewFieldBuilder().WithName("r").WithType(astbuilder.Selector("http", "Request").AsPointer(true))).
		AddResult(astbuilder.NewFieldBuilder().WithType(astbuilder.Selector(g.GetCurrentModelsPackage(), baseName+"Headers").AsPointer(true))).
		AddResult(astbuilder.ErrorField()).
		WithBody(bodyBuilder)

	g.HandlersFile.restBuilders = append(g.HandlersFile.restBuilders, fn)

	return nil
}

func (g *Generator) AddParseCookiesMethod(baseName string, params openapi3.Parameters) error {
	bodyBuilder := astbuilder.NewBodyBuilder().
		AddStmt(astbuilder.DeclareVar("cookies", astbuilder.Sel(astbuilder.I(g.GetCurrentModelsPackage()),   baseName+"Cookies").Build()))

	for _, param := range params {
		if param.Value.Schema == nil || param.Value.Schema.Value == nil {
			continue
		}

		varName := GoIdentLowercase(FormatGoLikeIdentifier(param.Value.Name))
		bodyBuilder.AddStmt(astbuilder.DefineCallWithErr(varName, astbuilder.Sel(astbuilder.I("r"),   "Cookie"), astbuilder.Str(param.Value.Name)))

		if param.Value.Required {
			bodyBuilder.AddStmt(astbuilder.IfErrNotNilReturn(astbuilder.I("nil").Build()))
		} else {
			bodyBuilder.AddStmt(astbuilder.If(astbuilder.And(astbuilder.Ne(astbuilder.I("err"), astbuilder.I("nil")), astbuilder.Not(astbuilder.Call(astbuilder.Sel(astbuilder.I("errors"),  "Is"), astbuilder.I("err"), astbuilder.Sel(astbuilder.I("http"),  "ErrNoCookie")))).Build()).WithBody(astbuilder.NewBodyBuilder().AddStmt(astbuilder.Return2(astbuilder.I("nil").Build(),  astbuilder.I("err").Build()))))
			g.AddHandlersImport("github.com/go-faster/errors")
		}

		if param.Value.Required {
			bodyBuilder.AddStmt(astbuilder.Define(astbuilder.I(varName+"Value").Build(),  astbuilder.Sel(astbuilder.I(varName),  "Value").Build()))

			switch {
			case param.Value.Schema.Value.Type.Permits("string"):
				g.AssignStringField(bodyBuilder, "cookies", varName+"Value", FormatGoLikeIdentifier(param.Value.Name), param.Value.Schema, param.Value.Required)
			default:
				return errors.New("unsupported path parameter type: " + fmt.Sprint(param.Value.Schema.Value.Type))
			}
		} else {
			ifBody := astbuilder.NewBodyBuilder().
				AddStmt(astbuilder.Define(astbuilder.I(varName+"Value").Build(),  astbuilder.Sel(astbuilder.I(varName),  "Value").Build()))
			g.AssignStringField(ifBody, "cookies", varName+"Value", FormatGoLikeIdentifier(param.Value.Name), param.Value.Schema, param.Value.Required)
			bodyBuilder.AddStmt(astbuilder.If(astbuilder.Eq(astbuilder.I("err"), astbuilder.I("nil")).Build()).WithBody(ifBody))
		}
	}

	bodyBuilder.
		AddStmt(astbuilder.Assign(astbuilder.I("err").Build(),  astbuilder.Call(astbuilder.Sel(astbuilder.Sel(astbuilder.I("h"),  "validator"),  "Struct"), astbuilder.I("cookies")).Build())).
		AddStmt(astbuilder.IfErrNotNilReturn(astbuilder.I("nil").Build())).
		AddStmt(astbuilder.Return2(astbuilder.Amp(astbuilder.Ident("cookies")).Build(),  astbuilder.I("nil").Build()))

	fn := astbuilder.NewFunctionBuilder().
		WithName("parse"+baseName+"Cookies").
		WithPointerReceiver("h", "Handler").
		AddParam(astbuilder.NewFieldBuilder().WithName("r").WithType(astbuilder.Selector("http", "Request").AsPointer(true))).
		AddResult(astbuilder.NewFieldBuilder().WithType(astbuilder.Selector(g.GetCurrentModelsPackage(), baseName+"Cookies").AsPointer(true))).
		AddResult(astbuilder.ErrorField()).
		WithBody(bodyBuilder)

	g.HandlersFile.restBuilders = append(g.HandlersFile.restBuilders, fn)

	return nil
}

func (g *Generator) AddParseRequestBodyMethod(baseName string, contentType string, body *openapi3.RequestBodyRef) error {
	bodyBuilder := astbuilder.NewBodyBuilder()

	if !body.Value.Required {
		bodyBuilder.AddStmt(astbuilder.IfNil(astbuilder.Sel(astbuilder.I("r"),  "Body").Build()).WithBody(astbuilder.NewBodyBuilder().
			AddStmt(astbuilder.Return2(astbuilder.I("nil").Build(),  astbuilder.I("nil").Build()))))
	}

	typeName := baseName + "RequestBody"
	var bodyType astbuilder.TypeExpressionBuilder
	content, ok := body.Value.Content[contentType]
	bodyType = astbuilder.Selector(g.GetCurrentModelsPackage(), typeName)
	if ok && content.Schema != nil {
		if content.Schema.Ref != "" {
			var importPath string
			typeName, importPath = g.ParseRefTypeName(content.Schema.Ref)
			bodyType = astbuilder.Selector(g.GetCurrentModelsPackage(), typeName)
			if importPath != "" {
				g.AddHandlersImport(importPath)
			}
			if refIsExternal(content.Schema.Ref) {
				bodyType = astbuilder.Ident(typeName)
			}
		}
	}

	bodyBuilder.AddStmt(astbuilder.DeclareVarWithType("bodyJSON", astbuilder.Selector("json", "RawMessage")))
	g.AddHandlersImport("encoding/json")

	bodyBuilder.
		AddStmt(astbuilder.DefineCall("err", astbuilder.Sel(astbuilder.Call(astbuilder.Sel(astbuilder.I("json"),   "NewDecoder"), astbuilder.Sel(astbuilder.I("r"),   "Body")),   "Decode"), astbuilder.Amp(astbuilder.Ident("bodyJSON")))).
		AddStmt(astbuilder.IfErrNotNilReturn(astbuilder.I("nil").Build())).
		AddStmt(astbuilder.Assign(astbuilder.I("err").Build(),  astbuilder.Call(g.GetValidateFuncStmt(typeName, content.Schema.Ref), astbuilder.I("bodyJSON")).Build())).
		AddStmt(astbuilder.IfErrNotNilReturn(astbuilder.I("nil").Build()))

	bodyBuilder.AddStmt(astbuilder.DeclareVar("body", bodyType.Build()))

	bodyBuilder.
		AddStmt(astbuilder.Assign(astbuilder.I("err").Build(),  astbuilder.Call(astbuilder.Sel(astbuilder.I("json"),  "Unmarshal"), astbuilder.I("bodyJSON"), astbuilder.Amp(astbuilder.Ident("body"))).Build())).
		AddStmt(astbuilder.IfErrNotNilReturn(astbuilder.I("nil").Build())).
		AddStmt(astbuilder.Assign(astbuilder.I("err").Build(),  astbuilder.Call(astbuilder.Sel(astbuilder.Sel(astbuilder.I("h"),  "validator"),  "Struct"), astbuilder.I("body")).Build())).
		AddStmt(astbuilder.IfErrNotNilReturn(astbuilder.I("nil").Build())).
		AddStmt(astbuilder.Return2(astbuilder.Amp(astbuilder.Ident("body")).Build(),  astbuilder.I("nil").Build()))

	fn := astbuilder.Function("parse"+baseName+"RequestBody").
		WithPointerReceiver("h", "Handler").
		AddParam(astbuilder.NewFieldBuilder().WithName("r").WithType(astbuilder.Selector("http", "Request").AsPointer(true))).
		AddResultExpr(astbuilder.Star(bodyType)).
		AddResultExpr(astbuilder.Error()).
		WithBody(bodyBuilder)

	g.HandlersFile.restBuilders = append(g.HandlersFile.restBuilders, fn)

	return nil
}

func (g *Generator) AddParseRequestMethod(baseName string, contentType string, pathParams openapi3.Parameters,
	queryParams openapi3.Parameters, headers openapi3.Parameters, cookieParams openapi3.Parameters,
	body *openapi3.RequestBodyRef,
) {
	bodyBuilder := astbuilder.NewBodyBuilder()
	elts := []ast.Expr{}

	if len(pathParams) > 0 {
		elts = append(elts, astbuilder.KeyValue(astbuilder.I("Path").Build(), astbuilder.Star(astbuilder.Ident("pathParams")).Build()))
		bodyBuilder.
			AddStmt(astbuilder.DefineCallWithErr("pathParams", astbuilder.Sel(astbuilder.I("h"),   "parse"+baseName+"PathParams"), astbuilder.I("r"))).
			AddStmt(astbuilder.IfErrNotNilReturn(astbuilder.I("nil").Build()))
	}
	if len(queryParams) > 0 {
		elts = append(elts, astbuilder.KeyValue(astbuilder.I("Query").Build(), astbuilder.Star(astbuilder.Ident("queryParams")).Build()))
		bodyBuilder.
			AddStmt(astbuilder.DefineCallWithErr("queryParams", astbuilder.Sel(astbuilder.I("h"),   "parse"+baseName+"QueryParams"), astbuilder.I("r"))).
			AddStmt(astbuilder.IfErrNotNilReturn(astbuilder.I("nil").Build()))
	}
	if len(headers) > 0 {
		elts = append(elts, astbuilder.KeyValue(astbuilder.I("Headers").Build(), astbuilder.Star(astbuilder.Ident("headers")).Build()))
		bodyBuilder.
			AddStmt(astbuilder.DefineCallWithErr("headers", astbuilder.Sel(astbuilder.I("h"),   "parse"+baseName+"Headers"), astbuilder.I("r"))).
			AddStmt(astbuilder.IfErrNotNilReturn(astbuilder.I("nil").Build()))
	}
	if len(cookieParams) > 0 {
		elts = append(elts, astbuilder.KeyValue(astbuilder.I("Cookies").Build(), astbuilder.Star(astbuilder.Ident("cookieParams")).Build()))
		bodyBuilder.
			AddStmt(astbuilder.DefineCallWithErr("cookieParams", astbuilder.Sel(astbuilder.I("h"),   "parse"+baseName+"Cookies"), astbuilder.I("r"))).
			AddStmt(astbuilder.IfErrNotNilReturn(astbuilder.I("nil").Build()))
	}
	if body != nil && body.Value != nil {
		content, ok := body.Value.Content[contentType]
		if ok && content.Schema != nil {
			if body.Value.Required {
				elts = append(elts, astbuilder.KeyValue(astbuilder.I("Body").Build(), astbuilder.Star(astbuilder.Ident("body")).Build()))
			} else {
				elts = append(elts, astbuilder.KeyValue(astbuilder.I("Body").Build(), astbuilder.I("body").Build()))
			}
			bodyBuilder.
				AddStmt(astbuilder.DefineCallWithErr("body", astbuilder.Sel(astbuilder.I("h"),   "parse"+baseName+"RequestBody"), astbuilder.I("r"))).
				AddStmt(astbuilder.IfErrNotNilReturn(astbuilder.I("nil").Build()))
		}
	}

	bodyBuilder.AddStmt(astbuilder.Return2(
		astbuilder.Amp(astbuilder.CompositeLit(
			astbuilder.Selector(g.GetCurrentModelsPackage(), baseName+"Request"),
			elts...,
		)).Build(), 
		astbuilder.I("nil").Build()))

	fn := astbuilder.NewFunctionBuilder().
		WithName("parse"+baseName+"Request").
		WithPointerReceiver("h", "Handler").
		AddParam(astbuilder.NewFieldBuilder().WithName("r").WithType(astbuilder.Selector("http", "Request").AsPointer(true))).
		AddResult(astbuilder.NewFieldBuilder().WithType(astbuilder.Selector(g.GetCurrentModelsPackage(), baseName+"Request").AsPointer(true))).
		AddResult(astbuilder.ErrorField()).
		WithBody(bodyBuilder)

	g.HandlersFile.restBuilders = append(g.HandlersFile.restBuilders, fn)
}
