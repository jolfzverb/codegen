package generator

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-faster/errors"
)

func (h *HandlersFile) AddParseQueryParamsMethod(baseName string, params openapi3.Parameters) error {
	bodyList := []ast.Stmt{
		&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{ast.NewIdent("queryParams")},
						Type: &ast.SelectorExpr{
							X:   ast.NewIdent("models"),
							Sel: ast.NewIdent(baseName + "QueryParams"),
						},
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
			Lhs: []ast.Expr{ast.NewIdent(varName)},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X: &ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.SelectorExpr{
									X:   ast.NewIdent("r"),
									Sel: ast.NewIdent("URL"),
								},
								Sel: ast.NewIdent("Query"),
							},
							Args: []ast.Expr{},
						},
						Sel: ast.NewIdent("Get"),
					},
					Args: []ast.Expr{
						&ast.BasicLit{
							Kind:  token.STRING,
							Value: fmt.Sprintf("%q", param.Value.Name),
						},
					},
				},
			},
		})
		if param.Value.Required {
			bodyList = append(bodyList, &ast.IfStmt{
				Cond: &ast.BinaryExpr{
					X:  ast.NewIdent(varName),
					Op: token.EQL,
					Y:  ast.NewIdent(`""`),
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{
								ast.NewIdent("nil"),
								ast.NewIdent(fmt.Sprintf("errors.New(%q)", param.Value.Name+" query param is required")),
							},
						},
					},
				},
			})
			h.AddImport("github.com/go-faster/errors")
			switch {
			case param.Value.Schema.Value.Type.Permits("string"):
				bodyList = append(bodyList,
					h.AssignStringField("queryParams", varName, FormatGoLikeIdentifier(param.Value.Name), param.Value.Schema, param.Value.Required)...,
				)
			default:
				return errors.New(fmt.Sprintf("unsupported path parameter type: %v", param.Value.Schema.Value.Type)) //nolint:revive
			}
		} else {
			bodyList = append(bodyList, &ast.IfStmt{
				Cond: &ast.BinaryExpr{
					X:  ast.NewIdent(varName),
					Op: token.NEQ,
					Y:  ast.NewIdent(`""`),
				},
				Body: &ast.BlockStmt{
					List: h.AssignStringField("queryParams", varName, FormatGoLikeIdentifier(param.Value.Name), param.Value.Schema, param.Value.Required),
				},
			})
		}
	}
	bodyList = append(bodyList, &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent("err")},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.SelectorExpr{
						X:   ast.NewIdent("h"),
						Sel: ast.NewIdent("validator"),
					},
					Sel: ast.NewIdent("Struct"),
				},
				Args: []ast.Expr{
					ast.NewIdent("queryParams"),
				},
			},
		},
	})
	bodyList = append(bodyList, &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  ast.NewIdent("err"),
			Op: token.NEQ,
			Y:  ast.NewIdent("nil"),
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						ast.NewIdent("nil"),
						ast.NewIdent("err"),
					},
				},
			},
		},
	})

	bodyList = append(bodyList,
		&ast.ReturnStmt{
			Results: []ast.Expr{
				&ast.UnaryExpr{
					Op: token.AND,
					X:  ast.NewIdent("queryParams"),
				},
				ast.NewIdent("nil"),
			},
		},
	)
	h.restDecls = append(h.restDecls, Func("parse"+baseName+"QueryParams",
		Field("h", Star(ast.NewIdent("Handler")), ""),
		[]*ast.Field{
			Field("r", Star(Sel(ast.NewIdent("http"), "Request")), ""),
		},
		[]*ast.Field{
			Field("", Star(Sel(ast.NewIdent("models"), baseName+"QueryParams")), ""),
			Field("", ast.NewIdent("error"), ""),
		},
		bodyList,
	))

	return nil
}

func (h *HandlersFile) AssignStringField(paramsName string, varName string, fieldName string, param *openapi3.SchemaRef, required bool) []ast.Stmt {
	if param.Value.Format == "date-time" {
		h.AddImport("time")
		var result []ast.Stmt
		result = append(result, &ast.AssignStmt{
			Lhs: []ast.Expr{
				ast.NewIdent("parsed" + fieldName),
				ast.NewIdent("err"),
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   ast.NewIdent("time"),
						Sel: ast.NewIdent("Parse"),
					},
					Args: []ast.Expr{
						&ast.SelectorExpr{
							X:   ast.NewIdent("time"),
							Sel: ast.NewIdent("RFC3339"),
						},
						ast.NewIdent(varName),
					},
				},
			},
		})
		result = append(result, &ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X:  ast.NewIdent("err"),
				Op: token.NEQ,
				Y:  ast.NewIdent("nil"),
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							ast.NewIdent("nil"),
							ast.NewIdent(fmt.Sprintf("errors.Wrap(err, %q)", fieldName+" is not a valid date-time format")),
						},
					},
				},
			},
		})
		var rhs ast.Expr
		if required && !h.requiredFieldsArePointers {
			rhs = ast.NewIdent("parsed" + fieldName)
		} else {
			rhs = &ast.UnaryExpr{
				Op: token.AND,
				X:  ast.NewIdent("parsed" + fieldName),
			}
		}

		return append(result, &ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.SelectorExpr{
					X:   ast.NewIdent(paramsName),
					Sel: ast.NewIdent(fieldName),
				},
			},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{rhs},
		})
	}
	var rhs ast.Expr
	if required && !h.requiredFieldsArePointers {
		rhs = ast.NewIdent(varName)
	} else {
		rhs = &ast.UnaryExpr{
			Op: token.AND,
			X:  ast.NewIdent(varName),
		}
	}

	return []ast.Stmt{&ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.SelectorExpr{
				X:   ast.NewIdent(paramsName),
				Sel: ast.NewIdent(fieldName),
			},
		},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{rhs},
	}}
}

func (h *HandlersFile) AddParseHeadersMethod(baseName string, params openapi3.Parameters) error {
	bodyList := []ast.Stmt{
		&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{ast.NewIdent("headers")},
						Type: &ast.SelectorExpr{
							X:   ast.NewIdent("models"),
							Sel: ast.NewIdent(baseName + "Headers"),
						},
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
			Lhs: []ast.Expr{ast.NewIdent(varName)},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X: &ast.SelectorExpr{
							X:   ast.NewIdent("r"),
							Sel: ast.NewIdent("Header"),
						},
						Sel: ast.NewIdent("Get"),
					},
					Args: []ast.Expr{
						&ast.BasicLit{
							Kind:  token.STRING,
							Value: fmt.Sprintf("%q", param.Value.Name),
						},
					},
				},
			},
		})
		if param.Value.Required {
			bodyList = append(bodyList, &ast.IfStmt{
				Cond: &ast.BinaryExpr{
					X:  ast.NewIdent(varName),
					Op: token.EQL,
					Y:  ast.NewIdent(`""`),
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{
								ast.NewIdent("nil"),
								ast.NewIdent(fmt.Sprintf("errors.New(%q)", param.Value.Name+" header is required")),
							},
						},
					},
				},
			})
			h.AddImport("github.com/go-faster/errors")
			switch {
			case param.Value.Schema.Value.Type.Permits("string"):
				bodyList = append(bodyList,
					h.AssignStringField("headers", varName, FormatGoLikeIdentifier(param.Value.Name), param.Value.Schema, param.Value.Required)...,
				)
			default:
				return errors.New(fmt.Sprintf("unsupported path parameter type: %v", param.Value.Schema.Value.Type)) //nolint:revive
			}
		} else {
			bodyList = append(bodyList, &ast.IfStmt{
				Cond: &ast.BinaryExpr{
					X:  ast.NewIdent(varName),
					Op: token.NEQ,
					Y:  ast.NewIdent(`""`),
				},
				Body: &ast.BlockStmt{
					List: h.AssignStringField("headers", varName, FormatGoLikeIdentifier(param.Value.Name), param.Value.Schema, param.Value.Required),
				},
			})
		}
	}
	bodyList = append(bodyList, &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent("err")},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.SelectorExpr{
						X:   ast.NewIdent("h"),
						Sel: ast.NewIdent("validator"),
					},
					Sel: ast.NewIdent("Struct"),
				},
				Args: []ast.Expr{
					ast.NewIdent("headers"),
				},
			},
		},
	})
	bodyList = append(bodyList, &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  ast.NewIdent("err"),
			Op: token.NEQ,
			Y:  ast.NewIdent("nil"),
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						ast.NewIdent("nil"),
						ast.NewIdent("err"),
					},
				},
			},
		},
	})
	bodyList = append(bodyList,
		&ast.ReturnStmt{
			Results: []ast.Expr{
				&ast.UnaryExpr{
					Op: token.AND,
					X:  ast.NewIdent("headers"),
				},
				ast.NewIdent("nil"),
			},
		},
	)
	h.restDecls = append(h.restDecls, Func("parse"+baseName+"Headers",
		Field("h", Star(ast.NewIdent("Handler")), ""),
		[]*ast.Field{
			Field("r", Star(Sel(ast.NewIdent("http"), "Request")), ""),
		},
		[]*ast.Field{
			Field("", Star(Sel(ast.NewIdent("models"), baseName+"Headers")), ""),
			Field("", ast.NewIdent("error"), ""),
		},
		bodyList,
	))

	return nil
}

func (h *HandlersFile) AddParseRequestBodyMethod(baseName string, contentType string, body *openapi3.RequestBodyRef) error {
	bodyList := []ast.Stmt{}
	if !body.Value.Required {
		bodyList = append(bodyList, &ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X:  &ast.SelectorExpr{X: ast.NewIdent("r"), Sel: ast.NewIdent("Body")},
				Op: token.EQL,
				Y:  ast.NewIdent("nil"),
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							ast.NewIdent("nil"),
							ast.NewIdent("nil"),
						},
					},
				},
			},
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
					Names: []*ast.Ident{ast.NewIdent("body")},
					Type: &ast.SelectorExpr{
						X:   ast.NewIdent("models"),
						Sel: ast.NewIdent(typeName),
					},
				},
			},
		},
	})

	bodyList = append(bodyList, &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent("err")},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("json"),
							Sel: ast.NewIdent("NewDecoder"),
						},
						Args: []ast.Expr{
							&ast.SelectorExpr{
								X:   ast.NewIdent("r"),
								Sel: ast.NewIdent("Body"),
							},
						},
					},
					Sel: ast.NewIdent("Decode"),
				},
				Args: []ast.Expr{
					&ast.UnaryExpr{
						Op: token.AND,
						X:  ast.NewIdent("body"),
					},
				},
			},
		},
	})

	bodyList = append(bodyList, &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  ast.NewIdent("err"),
			Op: token.NEQ,
			Y:  ast.NewIdent("nil"),
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						ast.NewIdent("nil"),
						ast.NewIdent("err"),
					},
				},
			},
		},
	})

	bodyList = append(bodyList, &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent("err")},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.SelectorExpr{
						X:   ast.NewIdent("h"),
						Sel: ast.NewIdent("validator"),
					},
					Sel: ast.NewIdent("Struct"),
				},
				Args: []ast.Expr{
					ast.NewIdent("body"),
				},
			},
		},
	})
	bodyList = append(bodyList, &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  ast.NewIdent("err"),
			Op: token.NEQ,
			Y:  ast.NewIdent("nil"),
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						ast.NewIdent("nil"),
						ast.NewIdent("err"),
					},
				},
			},
		},
	})
	bodyList = append(bodyList,
		&ast.ReturnStmt{
			Results: []ast.Expr{
				&ast.UnaryExpr{
					Op: token.AND,
					X:  ast.NewIdent("body"),
				},
				ast.NewIdent("nil"),
			},
		},
	)

	h.restDecls = append(h.restDecls, Func(
		"parse"+baseName+"RequestBody",
		Field("h", Star(ast.NewIdent("Handler")), ""),
		[]*ast.Field{
			Field("r", Star(Sel(ast.NewIdent("http"), "Request")), ""),
		},
		[]*ast.Field{
			Field("", Star(Sel(ast.NewIdent("models"), typeName)), ""),
			Field("", ast.NewIdent("error"), ""),
		},
		bodyList,
	))

	return nil
}

func (h *HandlersFile) AddParseRequestMethod(baseName string, contentType string, pathParams openapi3.Parameters,
	queryParams openapi3.Parameters, headers openapi3.Parameters, cookieParams openapi3.Parameters,
	body *openapi3.RequestBodyRef,
) {
	bodyList := []ast.Stmt{}
	elts := []ast.Expr{}
	if len(pathParams) > 0 {
		elts = append(elts, &ast.KeyValueExpr{
			Key: ast.NewIdent("Path"),
			Value: &ast.StarExpr{
				X: ast.NewIdent("pathParams"),
			},
		})
		bodyList = append(bodyList, &ast.AssignStmt{
			Lhs: []ast.Expr{
				ast.NewIdent("pathParams"),
				ast.NewIdent("err"),
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   ast.NewIdent("h"),
						Sel: ast.NewIdent("parse" + baseName + "PathParams"),
					},
					Args: []ast.Expr{
						ast.NewIdent("r"),
					},
				},
			},
		})
		bodyList = append(bodyList, &ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X:  ast.NewIdent("err"),
				Op: token.NEQ,
				Y:  ast.NewIdent("nil"),
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							ast.NewIdent("nil"),
							ast.NewIdent("err"),
						},
					},
				},
			},
		})
	}
	if len(queryParams) > 0 {
		elts = append(elts, &ast.KeyValueExpr{
			Key: ast.NewIdent("Query"),
			Value: &ast.StarExpr{
				X: ast.NewIdent("queryParams"),
			},
		})
		bodyList = append(bodyList, &ast.AssignStmt{
			Lhs: []ast.Expr{
				ast.NewIdent("queryParams"),
				ast.NewIdent("err"),
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   ast.NewIdent("h"),
						Sel: ast.NewIdent("parse" + baseName + "QueryParams"),
					},
					Args: []ast.Expr{
						ast.NewIdent("r"),
					},
				},
			},
		})
		bodyList = append(bodyList, &ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X:  ast.NewIdent("err"),
				Op: token.NEQ,
				Y:  ast.NewIdent("nil"),
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							ast.NewIdent("nil"),
							ast.NewIdent("err"),
						},
					},
				},
			},
		})
	}
	if len(headers) > 0 {
		elts = append(elts, &ast.KeyValueExpr{
			Key: ast.NewIdent("Headers"),
			Value: &ast.StarExpr{
				X: ast.NewIdent("headers"),
			},
		})
		bodyList = append(bodyList, &ast.AssignStmt{
			Lhs: []ast.Expr{
				ast.NewIdent("headers"),
				ast.NewIdent("err"),
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   ast.NewIdent("h"),
						Sel: ast.NewIdent("parse" + baseName + "Headers"),
					},
					Args: []ast.Expr{
						ast.NewIdent("r"),
					},
				},
			},
		})
		bodyList = append(bodyList, &ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X:  ast.NewIdent("err"),
				Op: token.NEQ,
				Y:  ast.NewIdent("nil"),
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							ast.NewIdent("nil"),
							ast.NewIdent("err"),
						},
					},
				},
			},
		})
	}
	if len(cookieParams) > 0 {
		elts = append(elts, &ast.KeyValueExpr{
			Key: ast.NewIdent("Cookies"),
			Value: &ast.StarExpr{
				X: ast.NewIdent("cookieParams"),
			},
		})
		bodyList = append(bodyList, &ast.AssignStmt{
			Lhs: []ast.Expr{
				ast.NewIdent("cookieParams"),
				ast.NewIdent("err"),
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   ast.NewIdent("h"),
						Sel: ast.NewIdent("parse" + baseName + "CookieParams"),
					},
					Args: []ast.Expr{
						ast.NewIdent("r"),
					},
				},
			},
		})
		bodyList = append(bodyList, &ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X:  ast.NewIdent("err"),
				Op: token.NEQ,
				Y:  ast.NewIdent("nil"),
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							ast.NewIdent("nil"),
							ast.NewIdent("err"),
						},
					},
				},
			},
		})
	}
	if body != nil && body.Value != nil {
		content, ok := body.Value.Content[contentType]
		if ok && content.Schema != nil {
			if body.Value.Required {
				elts = append(elts, &ast.KeyValueExpr{
					Key: ast.NewIdent("Body"),
					Value: &ast.StarExpr{
						X: ast.NewIdent("body"),
					},
				})
			} else {
				elts = append(elts, &ast.KeyValueExpr{
					Key:   ast.NewIdent("Body"),
					Value: ast.NewIdent("body"),
				})
			}
			bodyList = append(bodyList, &ast.AssignStmt{
				Lhs: []ast.Expr{
					ast.NewIdent("body"),
					ast.NewIdent("err"),
				},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("h"),
							Sel: ast.NewIdent("parse" + baseName + "RequestBody"),
						},
						Args: []ast.Expr{
							ast.NewIdent("r"),
						},
					},
				},
			})
			bodyList = append(bodyList, &ast.IfStmt{
				Cond: &ast.BinaryExpr{
					X:  ast.NewIdent("err"),
					Op: token.NEQ,
					Y:  ast.NewIdent("nil"),
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{
								ast.NewIdent("nil"),
								ast.NewIdent("err"),
							},
						},
					},
				},
			})
		}
	}

	bodyList = append(bodyList,
		&ast.ReturnStmt{
			Results: []ast.Expr{
				&ast.UnaryExpr{
					Op: token.AND,
					X: &ast.CompositeLit{
						Type: &ast.SelectorExpr{
							X:   ast.NewIdent("models"),
							Sel: ast.NewIdent(baseName + "Request"),
						},
						Elts: elts,
					},
				},
				ast.NewIdent("nil"),
			},
		},
	)

	h.restDecls = append(h.restDecls, Func(
		"parse"+baseName+"Request",
		Field("h", Star(ast.NewIdent("Handler")), ""),
		[]*ast.Field{
			Field("r", Star(Sel(ast.NewIdent("http"), "Request")), ""),
		},
		[]*ast.Field{
			Field("", Star(Sel(ast.NewIdent("models"), baseName+"Request")), ""),
			Field("", ast.NewIdent("error"), ""),
		},
		bodyList,
	))
}

func (h *HandlersFile) AddCreateResponseModel(baseName string, code string, response *openapi3.ResponseRef) error {
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
				Names: []*ast.Ident{ast.NewIdent("body")},
				Type: &ast.SelectorExpr{
					X:   ast.NewIdent("models"),
					Sel: ast.NewIdent(typeName),
				},
			})
			constructorArgs = append(constructorArgs, &ast.KeyValueExpr{
				Key:   ast.NewIdent("Body"),
				Value: ast.NewIdent("body"),
			})
		}
	}

	if len(response.Value.Headers) > 0 {
		arglist = append(arglist, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent("headers")},
			Type: &ast.SelectorExpr{
				X:   ast.NewIdent("models"),
				Sel: ast.NewIdent(baseName + "Response" + code + "Headers"),
			},
		})
		constructorArgs = append(constructorArgs, &ast.KeyValueExpr{
			Key:   ast.NewIdent("Headers"),
			Value: ast.NewIdent("headers"),
		})
	}

	h.restDecls = append(h.restDecls, Func(baseName+code+"Response",
		nil,
		arglist,
		[]*ast.Field{
			Field("", Star(Sel(ast.NewIdent("models"), baseName+"Response")), ""),
		},
		[]ast.Stmt{
			&ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.UnaryExpr{
						Op: token.AND,
						X: &ast.CompositeLit{
							Type: &ast.SelectorExpr{
								X:   ast.NewIdent("models"),
								Sel: ast.NewIdent(baseName + "Response"),
							},
							Elts: []ast.Expr{
								&ast.KeyValueExpr{
									Key: ast.NewIdent("StatusCode"),
									Value: &ast.BasicLit{
										Kind:  token.INT,
										Value: code,
									},
								},
								&ast.KeyValueExpr{
									Key: ast.NewIdent("Response" + code),
									Value: &ast.UnaryExpr{
										Op: token.AND,
										X: &ast.CompositeLit{
											Type: &ast.SelectorExpr{
												X:   ast.NewIdent("models"),
												Sel: ast.NewIdent(baseName + "Response" + code),
											},
											Elts: constructorArgs,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	))

	return nil
}
