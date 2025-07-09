package generator

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-faster/errors"
)

func (g *Generator) AddParseQueryParamsMethod(baseName string, params openapi3.Parameters) error {
	bodyList := []ast.Stmt{
		&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{I("queryParams")},
						Type:  Sel(I("models"), baseName+"QueryParams"),
					},
				},
			},
		},
	}
	for _, param := range params {
		if param.Value.Schema == nil || param.Value.Schema.Value == nil {
			continue
		}

		varName := GoIdentLowercase(FormatGoLikeIdentifier(param.Value.Name))
		bodyList = append(bodyList, &ast.AssignStmt{
			Lhs: []ast.Expr{I(varName)},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: Sel(&ast.CallExpr{
						Fun:  Sel(Sel(I("r"), "URL"), "Query"),
						Args: []ast.Expr{},
					}, "Get"),
					Args: []ast.Expr{Str(param.Value.Name)},
				},
			},
		})
		if param.Value.Required {
			bodyList = append(bodyList, &ast.IfStmt{
				Cond: Eq(I(varName), Str("")),
				Body: &ast.BlockStmt{
					List: []ast.Stmt{Ret2(I("nil"),
						&ast.CallExpr{
							Fun: Sel(I("errors"), "New"),
							Args: []ast.Expr{
								Str(param.Value.Name + " query param is required"),
							},
						},
					)},
				},
			})
			g.AddHandlersImport("github.com/go-faster/errors")
			switch {
			case param.Value.Schema.Value.Type.Permits("string"):
				bodyList = append(bodyList,
					g.AssignStringField("queryParams", varName, FormatGoLikeIdentifier(param.Value.Name), param.Value.Schema, param.Value.Required)...,
				)
			default:
				return errors.New(fmt.Sprintf("unsupported path parameter type: %v", param.Value.Schema.Value.Type)) //nolint:revive
			}
		} else {
			bodyList = append(bodyList, &ast.IfStmt{
				Cond: Ne(I(varName), Str("")),
				Body: &ast.BlockStmt{
					List: g.AssignStringField("queryParams", varName, FormatGoLikeIdentifier(param.Value.Name), param.Value.Schema, param.Value.Required),
				},
			})
		}
	}
	bodyList = append(bodyList, &ast.AssignStmt{
		Lhs: []ast.Expr{I("err")},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: Sel(Sel(I("h"), "validator"), "Struct"),
				Args: []ast.Expr{
					I("queryParams"),
				},
			},
		},
	})
	bodyList = append(bodyList, &ast.IfStmt{
		Cond: Ne(I("err"), I("nil")),
		Body: &ast.BlockStmt{List: []ast.Stmt{Ret2(I("nil"), I("err"))}},
	})

	bodyList = append(bodyList, Ret2(Amp(I("queryParams")), I("nil")))

	g.HandlersFile.restDecls = append(g.HandlersFile.restDecls, Func("parse"+baseName+"QueryParams",
		Field("h", Star(I("Handler")), ""),
		[]*ast.Field{
			Field("r", Star(Sel(I("http"), "Request")), ""),
		},
		[]*ast.Field{
			Field("", Star(Sel(I("models"), baseName+"QueryParams")), ""),
			Field("", I("error"), ""),
		},
		bodyList,
	))

	return nil
}

func (g *Generator) AssignStringField(paramsName string, varName string, fieldName string, param *openapi3.SchemaRef, required bool) []ast.Stmt {
	if param.Value.Format == "date-time" {
		g.AddHandlersImport("time")
		var result []ast.Stmt
		result = append(result, &ast.AssignStmt{
			Lhs: []ast.Expr{
				I("parsed" + fieldName),
				I("err"),
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: Sel(I("time"), "Parse"),
					Args: []ast.Expr{
						Sel(I("time"), "RFC3339"),
						I(varName),
					},
				},
			},
		})
		result = append(result, &ast.IfStmt{
			Cond: Ne(I("err"), I("nil")),
			Body: &ast.BlockStmt{
				List: []ast.Stmt{Ret2(
					I("nil"),
					&ast.CallExpr{
						Fun: Sel(I("errors"), "Wrap"),
						Args: []ast.Expr{
							I("err"),
							Str(fieldName + " is not a valid date-time format"),
						},
					},
				)},
			},
		})
		var rhs ast.Expr
		if required && !g.HandlersFile.requiredFieldsArePointers {
			rhs = I("parsed" + fieldName)
		} else {
			rhs = Amp(I("parsed" + fieldName))
		}

		return append(result, &ast.AssignStmt{
			Lhs: []ast.Expr{Sel(I(paramsName), fieldName)},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{rhs},
		})
	}
	var rhs ast.Expr
	if required && !g.HandlersFile.requiredFieldsArePointers {
		rhs = I(varName)
	} else {
		rhs = Amp(I(varName))
	}

	return []ast.Stmt{&ast.AssignStmt{
		Lhs: []ast.Expr{Sel(I(paramsName), fieldName)},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{rhs},
	}}
}

func (g *Generator) AddParseHeadersMethod(baseName string, params openapi3.Parameters) error {
	bodyList := []ast.Stmt{
		&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{I("headers")},
						Type:  Sel(I("models"), baseName+"Headers"),
					},
				},
			},
		},
	}
	for _, param := range params {
		if param.Value.Schema == nil || param.Value.Schema.Value == nil {
			continue
		}
		varName := GoIdentLowercase(FormatGoLikeIdentifier(param.Value.Name))
		bodyList = append(bodyList, &ast.AssignStmt{
			Lhs: []ast.Expr{I(varName)},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun:  Sel(Sel(I("r"), "Header"), "Get"),
					Args: []ast.Expr{Str(param.Value.Name)},
				},
			},
		})
		if param.Value.Required {
			bodyList = append(bodyList, &ast.IfStmt{
				Cond: Eq(I(varName), Str("")),
				Body: &ast.BlockStmt{
					List: []ast.Stmt{Ret2(I("nil"),
						&ast.CallExpr{
							Fun: Sel(I("errors"), "New"),
							Args: []ast.Expr{
								Str(param.Value.Name + " header is required"),
							},
						},
					)},
				},
			})
			g.AddHandlersImport("github.com/go-faster/errors")
			switch {
			case param.Value.Schema.Value.Type.Permits("string"):
				bodyList = append(bodyList,
					g.AssignStringField("headers", varName, FormatGoLikeIdentifier(param.Value.Name),
						param.Value.Schema, param.Value.Required,
					)...,
				)
			default:
				return errors.New("unsupported path parameter type: " + fmt.Sprint(param.Value.Schema.Value.Type))
			}
		} else {
			bodyList = append(bodyList, &ast.IfStmt{
				Cond: Ne(I(varName), Str("")),
				Body: &ast.BlockStmt{
					List: g.AssignStringField("headers", varName, FormatGoLikeIdentifier(param.Value.Name),
						param.Value.Schema, param.Value.Required,
					),
				},
			})
		}
	}
	bodyList = append(bodyList, &ast.AssignStmt{
		Lhs: []ast.Expr{I("err")},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: Sel(Sel(I("h"), "validator"), "Struct"),
				Args: []ast.Expr{
					I("headers"),
				},
			},
		},
	})
	bodyList = append(bodyList, &ast.IfStmt{
		Cond: Ne(I("err"), I("nil")),
		Body: &ast.BlockStmt{List: []ast.Stmt{Ret2(I("nil"), I("err"))}},
	})
	bodyList = append(bodyList, Ret2(Amp(I("headers")), I("nil")))
	g.HandlersFile.restDecls = append(g.HandlersFile.restDecls, Func("parse"+baseName+"Headers",
		Field("h", Star(I("Handler")), ""),
		[]*ast.Field{
			Field("r", Star(Sel(I("http"), "Request")), ""),
		},
		[]*ast.Field{
			Field("", Star(Sel(I("models"), baseName+"Headers")), ""),
			Field("", I("error"), ""),
		},
		bodyList,
	))

	return nil
}

func (g *Generator) AddParseCookiesMethod(baseName string, params openapi3.Parameters) error {
	bodyList := []ast.Stmt{
		&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{I("cookies")},
						Type:  Sel(I("models"), baseName+"Cookies"),
					},
				},
			},
		},
	}
	for _, param := range params {
		if param.Value.Schema == nil || param.Value.Schema.Value == nil {
			continue
		}

		varName := GoIdentLowercase(FormatGoLikeIdentifier(param.Value.Name))
		bodyList = append(bodyList, &ast.AssignStmt{
			Lhs: []ast.Expr{I(varName), I("err")},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun:  Sel(I("r"), "Cookie"),
					Args: []ast.Expr{Str(param.Value.Name)},
				},
			},
		})

		if param.Value.Required {
			bodyList = append(bodyList, &ast.IfStmt{
				Cond: Ne(I("err"), I("nil")),
				Body: &ast.BlockStmt{List: []ast.Stmt{Ret2(I("nil"), I("err"))}},
			})
		} else {
			bodyList = append(bodyList, &ast.IfStmt{
				Cond: &ast.BinaryExpr{
					X:  Ne(I("err"), I("nil")),
					Op: token.LAND,
					Y: &ast.UnaryExpr{
						Op: token.NOT,
						X: &ast.CallExpr{
							Fun: Sel(I("errors"), "Is"),
							Args: []ast.Expr{
								I("err"),
								Sel(I("http"), "ErrNoCookie"),
							},
						},
					},
				},
				Body: &ast.BlockStmt{List: []ast.Stmt{Ret2(I("nil"), I("err"))}},
			})
			g.AddHandlersImport("github.com/go-faster/errors")
		}

		if param.Value.Required {
			bodyList = append(bodyList, &ast.AssignStmt{
				Lhs: []ast.Expr{I(varName + "Value")},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{Sel(I(varName), "Value")},
			})

			switch {
			case param.Value.Schema.Value.Type.Permits("string"):
				bodyList = append(bodyList,
					g.AssignStringField("cookies", varName+"Value", FormatGoLikeIdentifier(param.Value.Name),
						param.Value.Schema, param.Value.Required,
					)...,
				)
			default:
				return errors.New("unsupported path parameter type: " + fmt.Sprint(param.Value.Schema.Value.Type))
			}
		} else {
			ifBody := []ast.Stmt{&ast.AssignStmt{
				Lhs: []ast.Expr{I(varName + "Value")},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{Sel(I(varName), "Value")},
			}}
			ifBody = append(ifBody,
				g.AssignStringField("cookies", varName+"Value", FormatGoLikeIdentifier(param.Value.Name),
					param.Value.Schema, param.Value.Required,
				)...,
			)
			bodyList = append(bodyList, &ast.IfStmt{
				Cond: Eq(I("err"), I("nil")),
				Body: &ast.BlockStmt{
					List: ifBody,
				},
			})
		}
	}
	bodyList = append(bodyList, &ast.AssignStmt{
		Lhs: []ast.Expr{I("err")},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun:  Sel(Sel(I("h"), "validator"), "Struct"),
				Args: []ast.Expr{I("cookies")},
			},
		},
	})
	bodyList = append(bodyList, &ast.IfStmt{
		Cond: Ne(I("err"), I("nil")),
		Body: &ast.BlockStmt{List: []ast.Stmt{Ret2(I("nil"), I("err"))}},
	})
	bodyList = append(bodyList, Ret2(Amp(I("cookies")), I("nil")))
	g.HandlersFile.restDecls = append(g.HandlersFile.restDecls, Func("parse"+baseName+"Cookies",
		Field("h", Star(I("Handler")), ""),
		[]*ast.Field{
			Field("r", Star(Sel(I("http"), "Request")), ""),
		},
		[]*ast.Field{
			Field("", Star(Sel(I("models"), baseName+"Cookies")), ""),
			Field("", I("error"), ""),
		},
		bodyList,
	))

	return nil
}

func (g *Generator) AddParseRequestBodyMethod(baseName string, contentType string, body *openapi3.RequestBodyRef) error {
	bodyList := []ast.Stmt{}
	if !body.Value.Required {
		bodyList = append(bodyList, &ast.IfStmt{
			Cond: Eq(Sel(I("r"), "Body"), I("nil")),
			Body: &ast.BlockStmt{List: []ast.Stmt{Ret2(I("nil"), I("nil"))}},
		})
	}

	typeName := baseName + "RequestBody"

	content, ok := body.Value.Content[contentType]
	if ok && content.Schema != nil {
		if content.Schema.Ref != "" {
			typeName = ParseRefTypeName(content.Schema.Ref)
		}
	}

	bodyList = append(bodyList, &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{I("body")},
					Type:  Sel(I("models"), typeName),
				},
			},
		},
	})

	bodyList = append(bodyList, &ast.AssignStmt{
		Lhs: []ast.Expr{I("err")},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: Sel(
					&ast.CallExpr{
						Fun:  Sel(I("json"), "NewDecoder"),
						Args: []ast.Expr{Sel(I("r"), "Body")},
					},
					"Decode",
				),
				Args: []ast.Expr{
					Amp(I("body")),
				},
			},
		},
	})

	bodyList = append(bodyList, &ast.IfStmt{
		Cond: Ne(I("err"), I("nil")),
		Body: &ast.BlockStmt{List: []ast.Stmt{Ret2(I("nil"), I("err"))}},
	})

	bodyList = append(bodyList, &ast.AssignStmt{
		Lhs: []ast.Expr{I("err")},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: Sel(Sel(I("h"), "validator"), "Struct"),
				Args: []ast.Expr{
					I("body"),
				},
			},
		},
	})
	bodyList = append(bodyList, &ast.IfStmt{
		Cond: Ne(I("err"), I("nil")),
		Body: &ast.BlockStmt{List: []ast.Stmt{Ret2(I("nil"), I("err"))}},
	})
	bodyList = append(bodyList, Ret2(Amp(I("body")), I("nil")))

	g.HandlersFile.restDecls = append(g.HandlersFile.restDecls, Func(
		"parse"+baseName+"RequestBody",
		Field("h", Star(I("Handler")), ""),
		[]*ast.Field{
			Field("r", Star(Sel(I("http"), "Request")), ""),
		},
		[]*ast.Field{
			Field("", Star(Sel(I("models"), typeName)), ""),
			Field("", I("error"), ""),
		},
		bodyList,
	))

	return nil
}

func (g *Generator) AddParseRequestMethod(baseName string, contentType string, pathParams openapi3.Parameters,
	queryParams openapi3.Parameters, headers openapi3.Parameters, cookieParams openapi3.Parameters,
	body *openapi3.RequestBodyRef,
) {
	bodyList := []ast.Stmt{}
	elts := []ast.Expr{}
	if len(pathParams) > 0 {
		elts = append(elts, &ast.KeyValueExpr{
			Key:   I("Path"),
			Value: Star(I("pathParams")),
		})
		bodyList = append(bodyList, &ast.AssignStmt{
			Lhs: []ast.Expr{
				I("pathParams"),
				I("err"),
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: Sel(I("h"), "parse"+baseName+"PathParams"),
					Args: []ast.Expr{
						I("r"),
					},
				},
			},
		})
		bodyList = append(bodyList, &ast.IfStmt{
			Cond: Ne(I("err"), I("nil")),
			Body: &ast.BlockStmt{List: []ast.Stmt{Ret2(I("nil"), I("err"))}},
		})
	}
	if len(queryParams) > 0 {
		elts = append(elts, &ast.KeyValueExpr{
			Key:   I("Query"),
			Value: Star(I("queryParams")),
		})
		bodyList = append(bodyList, &ast.AssignStmt{
			Lhs: []ast.Expr{
				I("queryParams"),
				I("err"),
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: Sel(I("h"), "parse"+baseName+"QueryParams"),
					Args: []ast.Expr{
						I("r"),
					},
				},
			},
		})
		bodyList = append(bodyList, &ast.IfStmt{
			Cond: Ne(I("err"), I("nil")),
			Body: &ast.BlockStmt{List: []ast.Stmt{Ret2(I("nil"), I("err"))}},
		})
	}
	if len(headers) > 0 {
		elts = append(elts, &ast.KeyValueExpr{
			Key:   I("Headers"),
			Value: Star(I("headers")),
		})
		bodyList = append(bodyList, &ast.AssignStmt{
			Lhs: []ast.Expr{
				I("headers"),
				I("err"),
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: Sel(I("h"), "parse"+baseName+"Headers"),
					Args: []ast.Expr{
						I("r"),
					},
				},
			},
		})
		bodyList = append(bodyList, &ast.IfStmt{
			Cond: Ne(I("err"), I("nil")),
			Body: &ast.BlockStmt{List: []ast.Stmt{Ret2(I("nil"), I("err"))}},
		})
	}
	if len(cookieParams) > 0 {
		elts = append(elts, &ast.KeyValueExpr{
			Key:   I("Cookies"),
			Value: Star(I("cookieParams")),
		})
		bodyList = append(bodyList, &ast.AssignStmt{
			Lhs: []ast.Expr{
				I("cookieParams"),
				I("err"),
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: Sel(I("h"), "parse"+baseName+"Cookies"),
					Args: []ast.Expr{
						I("r"),
					},
				},
			},
		})
		bodyList = append(bodyList, &ast.IfStmt{
			Cond: Ne(I("err"), I("nil")),
			Body: &ast.BlockStmt{List: []ast.Stmt{Ret2(I("nil"), I("err"))}},
		})
	}
	if body != nil && body.Value != nil {
		content, ok := body.Value.Content[contentType]
		if ok && content.Schema != nil {
			if body.Value.Required {
				elts = append(elts, &ast.KeyValueExpr{
					Key:   I("Body"),
					Value: Star(I("body")),
				})
			} else {
				elts = append(elts, &ast.KeyValueExpr{
					Key:   I("Body"),
					Value: I("body"),
				})
			}
			bodyList = append(bodyList, &ast.AssignStmt{
				Lhs: []ast.Expr{
					I("body"),
					I("err"),
				},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: Sel(I("h"), "parse"+baseName+"RequestBody"),
						Args: []ast.Expr{
							I("r"),
						},
					},
				},
			})
			bodyList = append(bodyList, &ast.IfStmt{
				Cond: Ne(I("err"), I("nil")),
				Body: &ast.BlockStmt{List: []ast.Stmt{Ret2(I("nil"), I("err"))}},
			})
		}
	}

	bodyList = append(bodyList,
		Ret2(Amp(&ast.CompositeLit{
			Type: Sel(I("models"), baseName+"Request"),
			Elts: elts,
		}),
			I("nil"),
		),
	)

	g.HandlersFile.restDecls = append(g.HandlersFile.restDecls, Func(
		"parse"+baseName+"Request",
		Field("h", Star(I("Handler")), ""),
		[]*ast.Field{
			Field("r", Star(Sel(I("http"), "Request")), ""),
		},
		[]*ast.Field{
			Field("", Star(Sel(I("models"), baseName+"Request")), ""),
			Field("", I("error"), ""),
		},
		bodyList,
	))
}

func (g *Generator) AddCreateResponseModel(baseName string, code string, response *openapi3.ResponseRef) error {
	arglist := []*ast.Field{}
	constructorArgs := []ast.Expr{}

	if len(response.Value.Content) > 0 {
		// assume there is a json body
		json, ok := response.Value.Content["application/json"]
		if !ok {
			return errors.New("response content type 'application/json' not found")
		}
		if json.Schema != nil {
			typeName := baseName + "Response" + code + "Body"
			if json.Schema.Ref != "" {
				typeName = ParseRefTypeName(json.Schema.Ref)
			}
			arglist = append(arglist, &ast.Field{
				Names: []*ast.Ident{I("body")},
				Type:  Sel(I("models"), typeName),
			})
			constructorArgs = append(constructorArgs, &ast.KeyValueExpr{
				Key:   I("Body"),
				Value: I("body"),
			})
		}
	}

	if len(response.Value.Headers) > 0 {
		arglist = append(arglist, &ast.Field{
			Names: []*ast.Ident{I("headers")},
			Type:  Sel(I("models"), baseName+"Response"+code+"Headers"),
		})
		constructorArgs = append(constructorArgs, &ast.KeyValueExpr{
			Key:   I("Headers"),
			Value: I("headers"),
		})
	}

	g.HandlersFile.restDecls = append(g.HandlersFile.restDecls, Func(baseName+code+"Response",
		nil,
		arglist,
		[]*ast.Field{
			Field("", Star(Sel(I("models"), baseName+"Response")), ""),
		},
		[]ast.Stmt{Ret1(
			Amp(&ast.CompositeLit{
				Type: Sel(I("models"), baseName+"Response"),
				Elts: []ast.Expr{
					&ast.KeyValueExpr{
						Key: I("StatusCode"),
						Value: &ast.BasicLit{
							Kind:  token.INT,
							Value: code,
						},
					},
					&ast.KeyValueExpr{
						Key: I("Response" + code),
						Value: Amp(&ast.CompositeLit{
							Type: Sel(I("models"), baseName+"Response"+code),
							Elts: constructorArgs,
						}),
					},
				},
			}),
		)},
	))

	return nil
}
